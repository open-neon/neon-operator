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

package safekeeper

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

// Operator manages lifecycle for SafeKeeper resources.
type Operator struct {
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// New creates a new SafeKeeper Operator.
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

// sync reconciles the SafeKeeper resource state with the desired state.
func (o *Operator) sync(ctx context.Context, name, namespace string) error {

	sk := &v1alpha1.SafeKeeper{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, sk); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	sk = sk.DeepCopy()

	key := fmt.Sprintf("%s/%s", namespace, name)

	logger := o.logger.With("key", key)
	logger.Info("syncing safekeeper")

	profile := &v1alpha1.SafeKeeperProfile{}
	if err := o.nclient.Get(ctx, client.ObjectKey{
		Name:      sk.Spec.ProfileRef.Name,
		Namespace: sk.Spec.ProfileRef.Namespace,
	}, profile); err != nil {
		return fmt.Errorf("failed to get safekeeper profile : %w", err)
	}

	profile = profile.DeepCopy()

	if err := o.updateStatefulSet(ctx, sk, profile); err != nil {
		return fmt.Errorf("failed to reconcile safekeeper statefulset: %w", err)
	}

	return nil
}

func (o *Operator) updateStatefulSet(ctx context.Context, sk *v1alpha1.SafeKeeper, profile *v1alpha1.SafeKeeperProfile) error {
	ss, err := o.kclient.AppsV1().StatefulSets(sk.GetNamespace()).Get(ctx, sk.GetName(), metav1.GetOptions{})
	notFound := false
	if err != nil {
		if apierrors.IsNotFound(err) {
			notFound = true
		} else {
			return fmt.Errorf("failed to get safekeeper statefulset: %w", err)
		}
	}

	spec, err := makeSafeKeeperStatefulSetSpec(profile)
	if err != nil {
		return fmt.Errorf("failed to create safekeeper statefulset spec: %w", err)
	}
	sset, err := makeSafeKeeperStatefulSet(sk, profile, spec)
	if err != nil {
		return fmt.Errorf("failed to create safekeeper statefulset object: %w", err)
	}
	hash, err := k8sutils.CreateInputHash(sk.ObjectMeta, spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for safekeeper statefulset: %w", err)
	}

	if notFound {
		if ss.Annotations == nil {
			ss.Annotations = make(map[string]string)
		}
		ss.Annotations[k8sutils.InputHashAnnotationKey] = hash

		_, err = o.kclient.AppsV1().StatefulSets(sk.GetNamespace()).Create(ctx, ss, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create safekeeper statefulset: %w", err)
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

	_, err = o.kclient.AppsV1().StatefulSets(sk.GetNamespace()).Update(ctx, ss, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update safekeeper statefulset: %w", err)
	}

	return nil
}
