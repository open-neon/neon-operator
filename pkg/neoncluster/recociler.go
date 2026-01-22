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

	corev1alpha1 "github.com/stateless-pg.io/neon-operator/pkg/api/v1alpha1"
)

const controllerName = "neoncluster-controller"

// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=neonclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=neonclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=neonclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=pageserverprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=pageserverprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=safekeeperprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=safekeeperprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=storagebrokerprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=storagebrokerprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=pageservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=pageservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=pageservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=safekeepers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=safekeepers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=safekeepers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=storagebrokers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=storagebrokers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.stateless-pg.io.io,resources=storagebrokers/finalizers,verbs=update

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
		Named("neoncluster").
		Complete(r)
}
