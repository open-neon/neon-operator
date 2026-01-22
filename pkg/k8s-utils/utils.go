package k8sutils

import (
	"os"
)

const (
	// OPERATOR_NAMESPACE is the environment variable name for the operator's namespace
	OPERATOR_NAMESPACE = "OPERATOR_NAMESPACE"
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
