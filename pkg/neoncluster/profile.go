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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/open-neon/neon-operator/pkg/api/v1alpha1"
	"github.com/open-neon/neon-operator/pkg/k8s-utils"
)

const (
	DefaultPageServerProfileName   = "default-pageserver"
	DefaultSafeKeeperProfileName   = "default-safekeeper"
	DefaultStorageBrokerProfileName = "default-storage-broker"
)

// getProfiles fetches all referenced profiles from the NeonCluster spec
// If a profile is not explicitly referenced, it will attempt to fetch the default profile
// Returns an error if any referenced or default profile does not exist
func (r *Operator) getProfiles(ctx context.Context, nc *corev1alpha1.NeonCluster) (*Profiles, error) {
	profiles := &Profiles{}

	pageServerProfileName := DefaultPageServerProfileName
	pageServerNamespace := k8sutils.GetOperatorNamespace()
	if nc.Spec.PageServerProfileRef != nil {
		pageServerProfileName = nc.Spec.PageServerProfileRef.Name
		pageServerNamespace = nc.Spec.PageServerProfileRef.Namespace
	}
	profile := &corev1alpha1.PageServerProfile{}
	if err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      pageServerProfileName,
		Namespace: pageServerNamespace,
	}, profile); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("PageServerProfile %s/%s not found", pageServerNamespace, pageServerProfileName)
		}
		return nil, fmt.Errorf("failed to get PageServerProfile %s/%s: %w", pageServerNamespace, pageServerProfileName, err)
	}
	profiles.pageServer = profile.DeepCopy()

	safeKeeperProfileName := DefaultSafeKeeperProfileName
	safeKeeperNamespace := k8sutils.GetOperatorNamespace()
	if nc.Spec.SafeKeeperProfileRef != nil {
		safeKeeperProfileName = nc.Spec.SafeKeeperProfileRef.Name
		safeKeeperNamespace = nc.Spec.SafeKeeperProfileRef.Namespace
	}
	skProfile := &corev1alpha1.SafeKeeperProfile{}
	if err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      safeKeeperProfileName,
		Namespace: safeKeeperNamespace,
	}, skProfile); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("SafeKeeperProfile %s/%s not found", safeKeeperNamespace, safeKeeperProfileName)
		}
		return nil, fmt.Errorf("failed to get SafeKeeperProfile %s/%s: %w", safeKeeperNamespace, safeKeeperProfileName, err)
	}
	profiles.safeKeeper = skProfile.DeepCopy()

	storageBrokerProfileName := DefaultStorageBrokerProfileName
	storageBrokerNamespace := k8sutils.GetOperatorNamespace()
	if nc.Spec.StorageBrokerProfileRef != nil {
		storageBrokerProfileName = nc.Spec.StorageBrokerProfileRef.Name
		storageBrokerNamespace = nc.Spec.StorageBrokerProfileRef.Namespace
	}
	sbProfile := &corev1alpha1.StorageBrokerProfile{}
	if err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      storageBrokerProfileName,
		Namespace: storageBrokerNamespace,
	}, sbProfile); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("StorageBrokerProfile %s/%s not found", storageBrokerNamespace, storageBrokerProfileName)
		}
		return nil, fmt.Errorf("failed to get StorageBrokerProfile %s/%s: %w", storageBrokerNamespace, storageBrokerProfileName, err)
	}
	profiles.storageBroker = sbProfile.DeepCopy()

	return profiles, nil
}
