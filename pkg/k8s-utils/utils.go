package k8sutils

import (
	"fmt"
	"os"

	"github.com/mitchellh/hashstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// OPERATOR_NAMESPACE is the environment variable name for the operator's namespace
	OPERATOR_NAMESPACE = "OPERATOR_NAMESPACE"
	// InputHashAnnotationKey is the annotation key for storing input hash
	InputHashAnnotationKey = "neon.io/input-hash"
)

// GetOperatorNamespace returns the namespace where the operator is running.
// It reads the OPERATOR_NAMESPACE environment variable that is injected via the Kubernetes Downward API.
// It panics if the environment variable is not set.
func GetOperatorNamespace() string {
	namespace := os.Getenv(OPERATOR_NAMESPACE)
	if namespace == "" {
		panic("OPERATOR_NAMESPACE environment variable is not set")
	}
	return namespace
}

func CreateInputHash(objMeta metav1.ObjectMeta, spec interface{}) (string, error) {
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
