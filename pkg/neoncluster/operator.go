package neoncluster

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mitchellh/hashstructure"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/open-neon/neon-operator/pkg/api/v1alpha1"
	corev1alpha1 "github.com/open-neon/neon-operator/pkg/api/v1alpha1"
)

// Operator manages lifecycle for NeonCluster resources.
type Operator struct {
	nclient client.Client
	kclient kubernetes.Interface
	scheme  *runtime.Scheme
	logger  *slog.Logger
}

// WorkloadObject is an interface that abstracts pageserver, safekeeper, and strogebroker objects.
type WorkloadObject interface {
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetGeneration() int64
}

// New creates a new NeonCluster Controller.
func New(client client.Client, scheme *runtime.Scheme, logger *slog.Logger) (*Operator, error) {
	logger = logger.With("component", controllerName)
	return &Operator{
		logger:  logger,
		nclient: client,
		scheme:  scheme,
	}, nil
}

// sync runes everytime where there is reconcile event for neocluster.
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

	err = r.updatePageServer(ctx, nc)
	return err
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

func (r *Operator) updatePageServer(ctx context.Context, nc *v1alpha1.NeonCluster) error {

	pg, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Get(ctx, fmt.Sprintf("%s-pageserver", nc.Name), metav1.GetOptions{})
	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get pageserver statefulset: %w", err)
	}

	sset, err := makePageServerStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset spec: %w", err)
	}

	hash, err := createInputHash(sset, sset.Spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for pageserver: %w", err)
	}

	// If StatefulSet doesn't exist, create it
	if notFound {
		newSS, err := makePageServerStatefulSet(nc)
		if err != nil {
			return fmt.Errorf("failed to create pageserver statefulset: %w", err)
		}

		newSS.Name = fmt.Sprintf("%s-pageserver", nc.Name)
		newSS.Namespace = nc.Namespace
		if newSS.Annotations == nil {
			newSS.Annotations = make(map[string]string)
		}
		newSS.Annotations[InputHashAnnotationKey] = hash

		if _, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Create(ctx, newSS, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create pageserver statefulset: %w", err)
		}

		r.logger.Info("pageserver statefulset created successfully")
		return nil
	}

	// Check if the input hash has changed
	if pg.Annotations[InputHashAnnotationKey] == hash {
		r.logger.Info("pageserver statefulset is up to date")
		return nil
	}

	// Create new StatefulSet with updated spec
	newSS, err := makePageServerStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset: %w", err)
	}

	// Preserve existing fields that shouldn't be overwritten
	newSS.Name = pg.Name
	newSS.Namespace = pg.Namespace
	newSS.ResourceVersion = pg.ResourceVersion
	newSS.UID = pg.UID
	if newSS.Annotations == nil {
		newSS.Annotations = make(map[string]string)
	}
	newSS.Annotations[InputHashAnnotationKey] = hash

	// Update the StatefulSet
	if _, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Update(ctx, newSS, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update pageserver statefulset: %w", err)
	}

	r.logger.Info("pageserver statefulset updated successfully")
	return nil
}

func createInputHash(obj WorkloadObject, spec interface{}) (string, error) {
	// Get all annotations and exclude the input hash annotation
	annotations := obj.GetAnnotations()
	filteredAnnotations := make(map[string]string)
	for k, v := range annotations {
		if k != InputHashAnnotationKey {
			filteredAnnotations[k] = v
		}
	}

	hash, err := hashstructure.Hash(struct {
		Labels      map[string]string
		Annotations map[string]string
		Generation  int64
		Spec        interface{}
	}{
		Labels:      obj.GetLabels(),
		Annotations: filteredAnnotations,
		Generation:  obj.GetGeneration(),
		Spec:        spec,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}
