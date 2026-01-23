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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
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

	return nil
}
