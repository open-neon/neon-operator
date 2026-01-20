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
	"log/slog"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/open-neon/neon-operator/pkg/api/v1alpha1"
)

const controllerName = "neoncluster-controller"

// Operator manages lifecycle for NeonCluster resources.
type Operator struct {
	nclient client.Client
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// New creates a new NeonCluster Controller.
func New(logger *slog.Logger, client client.Client, scheme *runtime.Scheme) (*Operator, error) {
	logger = logger.With("component", controllerName)
	return &Operator{
		logger:  logger,
		nclient: client,
		scheme:  scheme,
	}, nil
}

// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.open-neon.io,resources=neonclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NeonCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *Operator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Operator) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.NeonCluster{}).
		Named("neoncluster").
		Complete(r)
}
