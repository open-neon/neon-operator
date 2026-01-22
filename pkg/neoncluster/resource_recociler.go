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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1alpha1 "github.com/open-neon/neon-operator/pkg/api/v1alpha1"
)

const controllerName = "neoncluster-controller"

// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.open-neon.io,resources=pageserverprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.open-neon.io,resources=pageserverprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.open-neon.io,resources=safekeeperprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.open-neon.io,resources=safekeeperprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.open-neon.io,resources=storagebrokerprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.open-neon.io,resources=storagebrokerprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *Operator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	err := r.sync(ctx, req.Name, req.Namespace)
	return ctrl.Result{}, fmt.Errorf("Failed to sync neoncluster %s/%s: %w", req.Namespace, req.Name, err)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Operator) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.NeonCluster{}).
		Watches(&corev1alpha1.PageServerProfile{}, handler.EnqueueRequestsFromMapFunc(r.mapPageServerProfileToNeonCluster(mgr.GetClient()))).
		Watches(&corev1alpha1.SafeKeeperProfile{}, handler.EnqueueRequestsFromMapFunc(r.mapSafeKeeperProfileToNeonCluster(mgr.GetClient()))).
		Watches(&corev1alpha1.StorageBrokerProfile{}, handler.EnqueueRequestsFromMapFunc(r.mapStorageBrokerProfileToNeonCluster(mgr.GetClient()))).
		Named("neoncluster").
		Complete(r)
}

// mapPageServerProfileToNeonCluster returns NeonClusters that reference a PageServerProfile
func (r *Operator) mapPageServerProfileToNeonCluster(c client.Client) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		profile := o.(*corev1alpha1.PageServerProfile)
		var clusters corev1alpha1.NeonClusterList
		if err := c.List(ctx, &clusters, client.InNamespace(profile.Namespace)); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, cluster := range clusters.Items {
			if cluster.Spec.PageServerProfileRef != nil &&
				cluster.Spec.PageServerProfileRef.Name == profile.Name &&
				cluster.Spec.PageServerProfileRef.Namespace == profile.Namespace {
				requests = append(requests, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&cluster),
				})
			}
		}
		return requests
	}
}

// mapSafeKeeperProfileToNeonCluster returns NeonClusters that reference a SafeKeeperProfile
func (r *Operator) mapSafeKeeperProfileToNeonCluster(c client.Client) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		profile := o.(*corev1alpha1.SafeKeeperProfile)
		var clusters corev1alpha1.NeonClusterList
		if err := c.List(ctx, &clusters, client.InNamespace(profile.Namespace)); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, cluster := range clusters.Items {
			if cluster.Spec.SafeKeeperProfileRef != nil &&
				cluster.Spec.SafeKeeperProfileRef.Name == profile.Name &&
				cluster.Spec.SafeKeeperProfileRef.Namespace == profile.Namespace {
				requests = append(requests, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&cluster),
				})
			}
		}
		return requests
	}
}

// mapStorageBrokerProfileToNeonCluster returns NeonClusters that reference a StorageBrokerProfile
func (r *Operator) mapStorageBrokerProfileToNeonCluster(c client.Client) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		profile := o.(*corev1alpha1.StorageBrokerProfile)
		var clusters corev1alpha1.NeonClusterList
		if err := c.List(ctx, &clusters, client.InNamespace(profile.Namespace)); err != nil {
			return nil
		}

		var requests []reconcile.Request
		for _, cluster := range clusters.Items {
			if cluster.Spec.StorageBrokerProfileRef != nil &&
				cluster.Spec.StorageBrokerProfileRef.Name == profile.Name &&
				cluster.Spec.StorageBrokerProfileRef.Namespace == profile.Namespace {
				requests = append(requests, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&cluster),
				})
			}
		}
		return requests
	}
}
