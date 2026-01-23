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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1alpha1 "github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
)

const controllerName = "pageserver-controller"

// +kubebuilder:rbac:groups=core.stateless-pg.io,resources=pageservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io,resources=pageservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.stateless-pg.io,resources=pageservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io,resources=pageserverprofiles,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *Operator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if err := r.sync(ctx, req.Name, req.Namespace); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Operator) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.PageServer{}).
		Owns(&appsv1.StatefulSet{}).
		Watches(
			&corev1alpha1.PageServerProfile{},
			handler.EnqueueRequestsFromMapFunc(r.mapPageServerProfileToPageServers),
		).
		Named("pageserver").
		Complete(r)
}

// mapPageServerProfileToPageServers maps a PageServerProfile change to all PageServers that reference it.
func (r *Operator) mapPageServerProfileToPageServers(ctx context.Context, obj client.Object) []reconcile.Request {
	profile, ok := obj.(*corev1alpha1.PageServerProfile)
	if !ok {
		return []reconcile.Request{}
	}

	// List all PageServers across all namespaces
	pageservers := &corev1alpha1.PageServerList{}
	if err := r.nclient.List(ctx, pageservers); err != nil {
		r.logger.Error("failed to list pageservers", "error", err)
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0)
	for _, ps := range pageservers.Items {
		// Check if this PageServer references the changed profile
		if ps.Spec.ProfileRef != nil &&
			ps.Spec.ProfileRef.Name == profile.Name &&
			(ps.Spec.ProfileRef.Namespace == "" || ps.Spec.ProfileRef.Namespace == profile.Namespace) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      ps.Name,
					Namespace: ps.Namespace,
				},
			})
		}
	}

	return requests
}
