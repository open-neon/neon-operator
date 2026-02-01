/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pageserver

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	controlplane "github.com/stateless-pg/stateless-pg/pkg/control-plane"
	k8sutils "github.com/stateless-pg/stateless-pg/pkg/k8s-utils"
)

// Operator manages lifecycle for PageServer resources.
type Operator struct {
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// New creates a new PageServer Operator.
func New(nclient client.Client, scheme *runtime.Scheme, logger *slog.Logger, config *rest.Config) (*Operator, error) {
	logger = logger.With("component", controllerName)

	// Create kubernetes clientset for direct client-go operations
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Operator{
		logger:  logger,
		nclient: nclient,
		kclient: kclient,
		scheme:  scheme,
	}, nil
}

// sync reconciles the PageServer resource state with the desired state.
func (o *Operator) sync(ctx context.Context, name, namespace string) error {

	ps := &v1alpha1.PageServer{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, ps); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	ps = ps.DeepCopy()

	key := fmt.Sprintf("%s/%s", namespace, name)

	logger := o.logger.With("key", key)
	logger.Info("syncing pageserver")

	profile := &v1alpha1.PageServerProfile{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      ps.Spec.ProfileRef.Name,
		Namespace: ps.Spec.ProfileRef.Namespace,
	}, profile); err != nil {
		return fmt.Errorf("failed to get pageserver profile : %w", err)
	}

	profile = profile.DeepCopy()

	if err := o.updateHeadlessService(ctx, ps); err != nil {
		return fmt.Errorf("failed to reconcile pageserver headless service: %w", err)
	}

	if err := o.createPageServerConfigMap(ctx, ps, profile); err != nil {
		return fmt.Errorf("failed to create pageserver configmap: %w", err)
	}

	if err := o.updateStatefulSet(ctx, ps, profile); err != nil {
		return fmt.Errorf("failed to reconcile pageserver statefulset: %w", err)
	}

	return nil
}

func (o *Operator) updateStatefulSet(ctx context.Context, ps *v1alpha1.PageServer, profile *v1alpha1.PageServerProfile) error {
	ss, err := o.kclient.AppsV1().StatefulSets(ps.GetNamespace()).Get(ctx, ps.GetName(), metav1.GetOptions{})
	notFound := false
	if err != nil {
		if apierrors.IsNotFound(err) {
			notFound = true
		} else {
			return fmt.Errorf("failed to get pageserver statefulset: %w", err)
		}
	}

	spec, err := makePageServerStatefulSetSpec(ps, profile)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset spec: %w", err)
	}
	sset, err := makePageServerStatefulSet(ps, spec)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset object: %w", err)
	}
	hash, err := k8sutils.CreateInputHash(ps.ObjectMeta, spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for pageserver statefulset: %w", err)
	}

	if notFound {
		if ss.Annotations == nil {
			ss.Annotations = make(map[string]string)
		}
		ss.Annotations[k8sutils.InputHashAnnotationKey] = hash

		_, err = o.kclient.AppsV1().StatefulSets(ps.GetNamespace()).Create(ctx, ss, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create pageserver statefulset: %w", err)
		}
		return nil
	}

	if ss.Annotations[k8sutils.InputHashAnnotationKey] == hash {
		// No update needed
		return nil
	}

	ss.Spec = sset.Spec
	ss.Labels = sset.Labels
	if ss.Annotations == nil {
		ss.Annotations = make(map[string]string)
	}
	maps.Copy(ss.Annotations, sset.Annotations)
	ss.Annotations[k8sutils.InputHashAnnotationKey] = hash

	_, err = o.kclient.AppsV1().StatefulSets(ps.GetNamespace()).Update(ctx, ss, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update pageserver statefulset: %w", err)
	}

	return nil
}

func (o *Operator) updateHeadlessService(ctx context.Context, ps *v1alpha1.PageServer) error {
	svc, err := o.kclient.CoreV1().Services(ps.GetNamespace()).Get(ctx, ps.GetName(), metav1.GetOptions{})
	notFound := false
	if err != nil {
		if apierrors.IsNotFound(err) {
			notFound = true
			svc = &corev1.Service{}
		} else {
			return fmt.Errorf("failed to get pageserver service: %w", err)
		}
	}

	newSvc := makePageServerHeadlessService(ps)
	hash, err := k8sutils.CreateInputHash(ps.ObjectMeta, newSvc.Spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for pageserver service: %w", err)
	}

	if notFound {
		if newSvc.Annotations == nil {
			newSvc.Annotations = make(map[string]string)
		}
		newSvc.Annotations[k8sutils.InputHashAnnotationKey] = hash

		_, err = o.kclient.CoreV1().Services(ps.GetNamespace()).Create(ctx, newSvc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create pageserver service: %w", err)
		}
		return nil
	}

	if svc.Annotations[k8sutils.InputHashAnnotationKey] == hash {
		// No update needed
		return nil
	}

	svc.Spec = newSvc.Spec
	svc.Labels = newSvc.Labels
	if svc.Annotations == nil {
		svc.Annotations = make(map[string]string)
	}
	maps.Copy(svc.Annotations, newSvc.Annotations)
	svc.Annotations[k8sutils.InputHashAnnotationKey] = hash

	_, err = o.kclient.CoreV1().Services(ps.GetNamespace()).Update(ctx, svc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update pageserver service: %w", err)
	}

	return nil
}

func (o *Operator) createPageServerConfigMap(ctx context.Context, ps *v1alpha1.PageServer, psp *v1alpha1.PageServerProfile) error {
	configMapName := ps.GetName() + "-config"
	namespace := ps.GetNamespace()

	cm := &corev1.ConfigMap{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      configMapName,
		Namespace: namespace,
	}, cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get pageserver configmap: %w", err)
		}
		// Create new configmap
		tomlContent := generatePageServerToml(ps, psp)
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: namespace,
				Labels: map[string]string{
					"app":       "pageserver",
					"component": "pageserver-config",
				},
			},
			Data: map[string]string{
				"pageserver.toml": tomlContent,
			},
		}

		if err := o.nclient.Create(ctx, cm); err != nil {
			return fmt.Errorf("failed to create pageserver configmap: %w", err)
		}
		return nil
	}

	// Update existing configmap
	tomlContent := generatePageServerToml(ps, psp)
	cm.Data = map[string]string{
		"pageserver.toml": tomlContent,
	}

	if err := o.nclient.Update(ctx, cm); err != nil {
		return fmt.Errorf("failed to update pageserver configmap: %w", err)
	}

	return nil
}

func generatePageServerToml(ps *v1alpha1.PageServer, psp *v1alpha1.PageServerProfile) string {
	var sb strings.Builder

	// Control plane settings
	sb.WriteString(fmt.Sprintf("control_plane_api = '%s'\n", fmt.Sprintf("%s://%s.%s.svc.cluster.local:%s", controlplane.GetProtocol(), controlplane.ServiceName, k8sutils.GetOperatorNamespace(), controlplane.GetPort())))
	sb.WriteString(fmt.Sprintf("control_plane_emergency_mode = '%t'\n", psp.Spec.ControlPlane.EmergencyMode))

	neonClusterName := ps.Labels["neoncluster"]
	sb.WriteString(fmt.Sprintf("broker_endpoint = '%s'\n", fmt.Sprintf("http://%s-broker.%s.svc.cluster.local:50051", neonClusterName, ps.GetNamespace())))
	// Network settings
	sb.WriteString(fmt.Sprintf("listen_pg_addr = '%s'\n", "0.0.0.0:6400"))
	sb.WriteString(fmt.Sprintf("http_listen_addr = '%s'\n", "0.0.0.0:9898"))

	// TLS settings
	if psp.Spec.Security.EnableTLS {
		sb.WriteString(fmt.Sprintf("listen_https_addr = '%s'\n", "0.0.0.0:9899"))
		sb.WriteString(fmt.Sprintf("ssl_cert_file = '%s'\n", TLSCertPath))
		sb.WriteString(fmt.Sprintf("ssl_key_file = '%s'\n", TLSKeyPath))
	}

	sb.WriteString(fmt.Sprintf("checkpoint_distance = '%s'\n", psp.Spec.Durability.CheckpointDistance))

	sb.WriteString(fmt.Sprintf("checkpoint_timeout = '%s'\n", psp.Spec.Durability.CheckpointTimeout))

	sb.WriteString(fmt.Sprintf("gc_horizon = '%s'\n", psp.Spec.Retention.HistoryRetention))

	sb.WriteString(fmt.Sprintf("gc_period = '%s'\n", psp.Spec.Retention.GCInterval))

	sb.WriteString(fmt.Sprintf("pitr_interval = '%s'\n", psp.Spec.Retention.PITRRetention))

	sb.WriteString(fmt.Sprintf("ingest_batch_size = %d\n", *psp.Spec.Performance.IngestBatchSize))

	sb.WriteString(fmt.Sprintf("virtual_file_io_mode = %s\n", psp.Spec.Performance.IOMode))

	sb.WriteString(fmt.Sprintf("log_level = '%s'\n", strings.ToLower(psp.Spec.Observability.LogLevel)))

	// ssl_ca_certs = "/path/to/control-plane-selfsigned-cert.pem"
    // control_plane_api_token = "eyJ0eXAi..."

   // http_auth_type = "NeonJWT | Trust"       # Storage controller must send JWT
   // pg_auth_type = "NeonJWT"          # Compute nodes must send JWT
   // grpc_auth_type = "NeonJWT"        # Storage controller gRPC must use JWT
   // auth_validation_public_key_path = "/etc/neon/auth_public_key.pem"

	return sb.String()
}
