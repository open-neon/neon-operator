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
)

// getProfiles fetches all referenced profiles from the NeonCluster spec
// Returns an error if any referenced profile does not exist
func (r *Operator) getProfiles(ctx context.Context, nc *corev1alpha1.NeonCluster) (*Profiles, error) {
	profiles := &Profiles{}

	if nc.Spec.PageServerProfileRef != nil {
		profile := &corev1alpha1.PageServerProfile{}
		if err := r.nclient.Get(ctx, client.ObjectKey{
			Name:      nc.Spec.PageServerProfileRef.Name,
			Namespace: nc.Spec.PageServerProfileRef.Namespace,
		}, profile); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("PageServerProfile %s/%s not found", nc.Spec.PageServerProfileRef.Namespace, nc.Spec.PageServerProfileRef.Name)
			}
			return nil, fmt.Errorf("failed to get PageServerProfile %s/%s: %w", nc.Spec.PageServerProfileRef.Namespace, nc.Spec.PageServerProfileRef.Name, err)
		}
		profiles.pageServer = profile.DeepCopy()
	}

	if nc.Spec.SafeKeeperProfileRef != nil {
		profile := &corev1alpha1.SafeKeeperProfile{}
		if err := r.nclient.Get(ctx, client.ObjectKey{
			Name:      nc.Spec.SafeKeeperProfileRef.Name,
			Namespace: nc.Spec.SafeKeeperProfileRef.Namespace,
		}, profile); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("SafeKeeperProfile %s/%s not found", nc.Spec.SafeKeeperProfileRef.Namespace, nc.Spec.SafeKeeperProfileRef.Name)
			}
			return nil, fmt.Errorf("failed to get SafeKeeperProfile %s/%s: %w", nc.Spec.SafeKeeperProfileRef.Namespace, nc.Spec.SafeKeeperProfileRef.Name, err)
		}
		profiles.safeKeeper = profile.DeepCopy()
	}

	if nc.Spec.StorageBrokerProfileRef != nil {
		profile := &corev1alpha1.StorageBrokerProfile{}
		if err := r.nclient.Get(ctx, client.ObjectKey{
			Name:      nc.Spec.StorageBrokerProfileRef.Name,
			Namespace: nc.Spec.StorageBrokerProfileRef.Namespace,
		}, profile); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("StorageBrokerProfile %s/%s not found", nc.Spec.StorageBrokerProfileRef.Namespace, nc.Spec.StorageBrokerProfileRef.Name)
			}
			return nil, fmt.Errorf("failed to get StorageBrokerProfile %s/%s: %w", nc.Spec.StorageBrokerProfileRef.Namespace, nc.Spec.StorageBrokerProfileRef.Name, err)
		}
		profiles.storageBroker = profile.DeepCopy()
	}

	return profiles, nil
}
