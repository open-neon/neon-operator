package neoncluster

import (
	"context"
	"fmt"

	"github.com/mitchellh/hashstructure"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/open-neon/neon-operator/pkg/api/v1alpha1"
)

// WorkloadObject is an interface that abstracts pageserver, safekeeper, and strogebroker objects.
type WorkloadObject interface {
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetGeneration() int64
}

func (r *Operator) sync(ctx context.Context, name string, namespace string) error {
	nc, err := r.getNeonCluster(ctx, name, namespace)
	if err != nil {
		return err
	}

	if nc == nil {
		return nil
	}

	key := fmt.Sprintf("%s/%s", namespace, name)

	logger := r.logger.With("key", key)

	logger.Info("Sync neoncluster")

	return nil

}

func (r *Operator) getNeonCluster(ctx context.Context, name string, namespace string) (*corev1alpha1.NeonCluster, error) {
	nc := &corev1alpha1.NeonCluster{}
	err := r.nclient.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, nc)
	if err != nil {
		return nil, err
	}

	return nc.DeepCopy(), nil
}

func createInputHash(obj WorkloadObject) (string, error) {
	var spec interface{}

	// Extract spec based on object type
	if sset, ok := obj.(*appsv1.StatefulSet); ok {
		spec = sset.Spec
	} else if deploy, ok := obj.(*appsv1.Deployment); ok {
		spec = deploy.Spec
	} else {
		return "", fmt.Errorf("unsupported workload type")
	}

	hash, err := hashstructure.Hash(struct {
		Labels      map[string]string
		Annotations map[string]string
		Generation  int64
		Spec        interface{}
	}{
		Labels:      obj.GetLabels(),
		Annotations: obj.GetAnnotations(),
		Generation:  obj.GetGeneration(),
		Spec:        spec,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}
