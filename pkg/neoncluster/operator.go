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

	if !notFound {
		pg = pg.DeepCopy()
	}

	sset, err := makePageServerStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset spec: %w", err)
	}

	hash, err := createInputHash(sset.ObjectMeta, sset.Spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for pageserver: %w", err)
	}

	newSS, err := makePageServerStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create pageserver statefulset: %w", err)
	}

	// If StatefulSet doesn't exist, create it
	if notFound {

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

func (r *Operator) updateSafekeeper(ctx context.Context, nc *v1alpha1.NeonCluster) error {

	sk, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Get(ctx, fmt.Sprintf("%s-safekeeper", nc.Name), metav1.GetOptions{})
	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get safekeeper statefulset: %w", err)
	}

	if !notFound {
		sk = sk.DeepCopy()
	}

	sset, err := makeSafekeeperStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create safekeeper statefulset spec: %w", err)
	}

	hash, err := createInputHash(sset.ObjectMeta, sset.Spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for safekeeper: %w", err)
	}

	newSS, err := makeSafekeeperStatefulSet(nc)
	if err != nil {
		return fmt.Errorf("failed to create safekeeper statefulset: %w", err)
	}

	// If StatefulSet doesn't exist, create it
	if notFound {

		newSS.Name = fmt.Sprintf("%s-safekeeper", nc.Name)
		newSS.Namespace = nc.Namespace
		if newSS.Annotations == nil {
			newSS.Annotations = make(map[string]string)
		}
		newSS.Annotations[InputHashAnnotationKey] = hash

		if _, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Create(ctx, newSS, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create safekeeper statefulset: %w", err)
		}

		r.logger.Info("safekeeper statefulset created successfully")
		return nil
	}

	// Check if the input hash has changed
	if sk.Annotations[InputHashAnnotationKey] == hash {
		r.logger.Info("safekeeper statefulset is up to date")
		return nil
	}

	// Preserve existing fields that shouldn't be overwritten
	newSS.Name = sk.Name
	newSS.Namespace = sk.Namespace
	newSS.ResourceVersion = sk.ResourceVersion
	newSS.UID = sk.UID
	if newSS.Annotations == nil {
		newSS.Annotations = make(map[string]string)
	}
	newSS.Annotations[InputHashAnnotationKey] = hash

	// Update the StatefulSet
	if _, err := r.kclient.AppsV1().StatefulSets(nc.Namespace).Update(ctx, newSS, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update safekeeper statefulset: %w", err)
	}

	r.logger.Info("safekeeper statefulset updated successfully")
	return nil
}

func (r *Operator) updateStorageBroker(ctx context.Context, nc *v1alpha1.NeonCluster) error {

	sb, err := r.kclient.AppsV1().Deployments(nc.Namespace).Get(ctx, fmt.Sprintf("%s-storage-broker", nc.Name), metav1.GetOptions{})
	notFound := apierrors.IsNotFound(err)

	if err != nil && !notFound {
		return fmt.Errorf("failed to get storage broker deployment: %w", err)
	}

	if !notFound {
		sb = sb.DeepCopy()
	}

	deploy, err := makeSafekeeperDeployment(nc)
	if err != nil {
		return fmt.Errorf("failed to create storage broker deployment spec: %w", err)
	}

	hash, err := createInputHash(deploy.ObjectMeta, deploy.Spec)
	if err != nil {
		return fmt.Errorf("failed to create input hash for storage broker: %w", err)
	}

	newDeploy, err := makeSafekeeperDeployment(nc)
	if err != nil {
		return fmt.Errorf("failed to create storage broker deployment: %w", err)
	}

	// If Deployment doesn't exist, create it
	if notFound {

		newDeploy.Name = fmt.Sprintf("%s-storage-broker", nc.Name)
		newDeploy.Namespace = nc.Namespace
		if newDeploy.Annotations == nil {
			newDeploy.Annotations = make(map[string]string)
		}
		newDeploy.Annotations[InputHashAnnotationKey] = hash

		if _, err := r.kclient.AppsV1().Deployments(nc.Namespace).Create(ctx, newDeploy, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("failed to create storage broker deployment: %w", err)
		}

		r.logger.Info("storage broker deployment created successfully")
		return nil
	}

	// Check if the input hash has changed
	if sb.Annotations[InputHashAnnotationKey] == hash {
		r.logger.Info("storage broker deployment is up to date")
		return nil
	}

	// Preserve existing fields that shouldn't be overwritten
	newDeploy.Name = sb.Name
	newDeploy.Namespace = sb.Namespace
	newDeploy.ResourceVersion = sb.ResourceVersion
	newDeploy.UID = sb.UID
	if newDeploy.Annotations == nil {
		newDeploy.Annotations = make(map[string]string)
	}
	newDeploy.Annotations[InputHashAnnotationKey] = hash

	// Update the Deployment
	if _, err := r.kclient.AppsV1().Deployments(nc.Namespace).Update(ctx, newDeploy, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update storage broker deployment: %w", err)
	}

	r.logger.Info("storage broker deployment updated successfully")
	return nil
}

func createInputHash(objMeta metav1.ObjectMeta, spec interface{}) (string, error) {
	// Get all annotations and exclude the input hash annotation
	filteredAnnotations := make(map[string]string)
	for k, v := range objMeta.Annotations {
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
		Labels:      objMeta.Labels,
		Annotations: filteredAnnotations,
		Generation:  objMeta.Generation,
		Spec:        spec,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to calculate combined hash: %w", err)
	}

	return fmt.Sprintf("%d", hash), nil
}
