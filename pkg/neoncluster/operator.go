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

package neoncluster

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	corev1alpha1 "github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	"github.com/stateless-pg/stateless-pg/pkg/operator"
)

// Operator manages lifecycle for NeonCluster resources.
type Operator struct {
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// Profiles holds references to all profile resources for a NeonCluster
type Profiles struct {
	pageServer    *corev1alpha1.PageServerProfile
	safeKeeper    *corev1alpha1.SafeKeeperProfile
	storageBroker *corev1alpha1.StorageBrokerProfile
}

// New creates a new NeonCluster Controller.
func New(client client.Client, scheme *runtime.Scheme, logger *slog.Logger, config *rest.Config) (*Operator, error) {
	logger = logger.With("component", controllerName)

	// Create kubernetes clientset for direct client-go operations
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Operator{
		logger:  logger,
		nclient: client,
		kclient: kclient,
		scheme:  scheme,
	}, nil
}

// sync runes everytime where there is reconcile event for neocluster.
func (r *Operator) sync(ctx context.Context, name, namespace string) error {
	nc := &corev1alpha1.NeonCluster{}
	if err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, nc); err != nil {
		if apierrors.IsNotFound(err) {
			// NeonCluster resource not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return nil
		}
		return err
	}

	nc = nc.DeepCopy()

	key := fmt.Sprintf("%s/%s", namespace, name)
	logger := r.logger.With("key", key)

	logger.Info("Sync neoncluster")

	pf, err := r.getProfiles(ctx, nc)
	if err != nil {
		return err
	}

	if err := r.updatePageServer(ctx, nc, pf.pageServer, logger); err != nil {
		return err
	}

	if err := r.updateSafeKeeper(ctx, nc, pf.safeKeeper, logger); err != nil {
		return err
	}

	return r.updateStorageBroker(ctx, nc, pf.storageBroker, logger)

}

func (r *Operator) updatePageServer(ctx context.Context, nc *v1alpha1.NeonCluster, profile *v1alpha1.PageServerProfile, logger *slog.Logger) error {

	ps := &v1alpha1.PageServer{}
	err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      nc.Name,
		Namespace: nc.Namespace,
	}, ps)

	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get pageserver: %w", err)
	}

	if !notFound {
		ps = ps.DeepCopy()
	}

	if !notFound && ps.Spec.ProfileRef != nil &&
		ps.Spec.ProfileRef.Name == profile.Name &&
		ps.Spec.ProfileRef.Namespace == profile.Namespace {
		// No update needed
		return nil
	}

	if notFound {
		ps = &v1alpha1.PageServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nc.Name,
				Namespace: nc.Namespace,
			},
			Spec: v1alpha1.PageServerSpec{
				ProfileRef: &corev1.ObjectReference{
					Name:      profile.Name,
					Namespace: profile.Namespace,
				},
			},
		}

		operator.UpdateObject(ps,
			operator.WithLabels(map[string]string{
				"neoncluster": nc.Name,
				"app":         "pageserver",
			}),
			operator.WithOwner(nc),
		)

		err = r.nclient.Create(ctx, ps)
		if err != nil {
			return fmt.Errorf("failed to create pageserver: %w", err)
		}

		logger.Info("Created pageserver", "name", ps.Name, "namespace", ps.Namespace)

		return nil
	}

	ps.Spec.ProfileRef = &corev1.ObjectReference{
		Name:      profile.Name,
		Namespace: profile.Namespace,
	}

	err = r.nclient.Update(ctx, ps)
	if err != nil {
		return fmt.Errorf("failed to update pageserver: %w", err)
	}

	logger.Info("Updated pageserver", "name", ps.Name, "namespace", ps.Namespace)

	return nil
}

func (r *Operator) updateSafeKeeper(ctx context.Context, nc *v1alpha1.NeonCluster, profile *v1alpha1.SafeKeeperProfile, logger *slog.Logger) error {
	skname := nc.Name

	sk := &v1alpha1.SafeKeeper{}
	err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      skname,
		Namespace: nc.Namespace,
	}, sk)

	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get safekeeper: %w", err)
	}

	if !notFound {
		sk = sk.DeepCopy()
	}

	if !notFound && sk.Spec.ProfileRef != nil &&
		sk.Spec.ProfileRef.Name == profile.Name &&
		sk.Spec.ProfileRef.Namespace == profile.Namespace {
		// No update needed
		return nil
	}

	if notFound {
		sk = &v1alpha1.SafeKeeper{
			ObjectMeta: metav1.ObjectMeta{
				Name:      skname,
				Namespace: nc.Namespace,
			},
			Spec: v1alpha1.SafeKeeperSpec{
				ProfileRef: &corev1.ObjectReference{
					Name:      profile.Name,
					Namespace: profile.Namespace,
				},
			},
		}

		operator.UpdateObject(sk,
			operator.WithLabels(map[string]string{
				"neoncluster": nc.Name,
				"app":         "safekeeper",
			}),
			operator.WithOwner(nc),
		)

		err = r.nclient.Create(ctx, sk)
		if err != nil {
			return fmt.Errorf("failed to create safekeeper: %w", err)
		}

		logger.Info("Created safekeeper", "name", sk.Name, "namespace", sk.Namespace)

		return nil
	}

	sk.Spec.ProfileRef = &corev1.ObjectReference{
		Name:      profile.Name,
		Namespace: profile.Namespace,
	}

	err = r.nclient.Update(ctx, sk)
	if err != nil {
		return fmt.Errorf("failed to update safekeeper: %w", err)
	}

	logger.Info("Updated safekeeper", "name", sk.Name, "namespace", sk.Namespace)

	return nil
}

func (r *Operator) updateStorageBroker(ctx context.Context, nc *v1alpha1.NeonCluster, profile *v1alpha1.StorageBrokerProfile, logger *slog.Logger) error {
	sbname := nc.Name

	sb := &v1alpha1.StorageBroker{}
	err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      sbname,
		Namespace: nc.Namespace,
	}, sb)

	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get storagebroker: %w", err)
	}

	if !notFound {
		sb = sb.DeepCopy()
	}

	if !notFound && sb.Spec.ProfileRef != nil &&
		sb.Spec.ProfileRef.Name == profile.Name &&
		sb.Spec.ProfileRef.Namespace == profile.Namespace {
		// No update needed
		return nil
	}

	if notFound {
		sb = &v1alpha1.StorageBroker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sbname,
				Namespace: nc.Namespace,
			},
			Spec: v1alpha1.StorageBrokerSpec{
				ProfileRef: &corev1.ObjectReference{
					Name:      profile.Name,
					Namespace: profile.Namespace,
				},
			},
		}

		operator.UpdateObject(sb,
			operator.WithLabels(map[string]string{
				"neoncluster": nc.Name,
				"app":         "storagebroker",
			}),
			operator.WithOwner(nc),
		)

		err = r.nclient.Create(ctx, sb)
		if err != nil {
			return fmt.Errorf("failed to create storagebroker: %w", err)
		}

		logger.Info("Created storagebroker", "name", sb.Name, "namespace", sb.Namespace)

		return nil
	}

	sb.Spec.ProfileRef = &corev1.ObjectReference{
		Name:      profile.Name,
		Namespace: profile.Namespace,
	}

	err = r.nclient.Update(ctx, sb)
	if err != nil {
		return fmt.Errorf("failed to update storagebroker: %w", err)
	}

	logger.Info("Updated storagebroker", "name", sb.Name, "namespace", sb.Namespace)

	return nil
}
