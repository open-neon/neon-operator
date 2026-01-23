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

package storagebroker

import (
	"context"
	"fmt"
	"log/slog"
	"maps"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	k8sutils "github.com/stateless-pg/stateless-pg/pkg/k8s-utils"
)

// Operator manages lifecycle for StorageBroker resources.
type Operator struct {
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// New creates a new StorageBroker Operator.
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

// sync reconciles the StorageBroker resource state with the desired state.
func (o *Operator) sync(ctx context.Context, name, namespace string) error {

	sb := &v1alpha1.StorageBroker{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, sb); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	sb = sb.DeepCopy()

	key := fmt.Sprintf("%s/%s", namespace, name)

	logger := o.logger.With("key", key)
	logger.Info("syncing storagebroker")

	profile := &v1alpha1.StorageBrokerProfile{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      sb.Spec.ProfileRef.Name,
		Namespace: sb.Spec.ProfileRef.Namespace,
	}, profile); err != nil {
		return fmt.Errorf("failed to get storagebroker profile : %w", err)
	}

	profile = profile.DeepCopy()

	if err := o.updateDeployment(ctx, sb, profile); err != nil {
		return fmt.Errorf("failed to reconcile storagebroker deployment: %w", err)
	}

	return nil
}

func (o *Operator) updateDeployment(ctx context.Context, sb *v1alpha1.StorageBroker, profile *v1alpha1.StorageBrokerProfile) error {
	dep, err := o.kclient.AppsV1().Deployments(sb.GetNamespace()).Get(ctx, sb.GetName(), metav1.GetOptions{})
	notFound := false
	if err != nil {
		if apierrors.IsNotFound(err) {
			notFound = true
		} else {
			return fmt.Errorf("failed to get storagebroker deployment: %w", err)
		}
	}

	spec, err := makeStorageBrokerDeploymentSpec(sb, profile)
	if err != nil {
		return fmt.Errorf("failed to create storagebroker deployment spec: %w", err)
	}
	deployment, err := makeStorageBrokerDeployment(sb, profile, spec)
	if err != nil {
		return fmt.Errorf("failed to create storagebroker deployment object: %w", err)
	}
	hash, err := k8sutils.CreateInputHash(sb.ObjectMeta, spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for storagebroker deployment: %w", err)
	}

	if notFound {
		if dep.Annotations == nil {
			dep.Annotations = make(map[string]string)
		}
		dep.Annotations[k8sutils.InputHashAnnotationKey] = hash

		_, err = o.kclient.AppsV1().Deployments(sb.GetNamespace()).Create(ctx, dep, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create storagebroker deployment: %w", err)
		}
		return nil
	}

	if dep.Annotations[k8sutils.InputHashAnnotationKey] == hash {
		// No update needed
		return nil
	}

	dep.Spec = deployment.Spec
	dep.Labels = deployment.Labels
	if dep.Annotations == nil {
		dep.Annotations = make(map[string]string)
	}
	maps.Copy(dep.Annotations, deployment.Annotations)
	dep.Annotations[k8sutils.InputHashAnnotationKey] = hash

	_, err = o.kclient.AppsV1().Deployments(sb.GetNamespace()).Update(ctx, dep, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update storagebroker deployment: %w", err)
	}

	return nil
}
