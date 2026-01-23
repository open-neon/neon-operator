package operator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ManagedByLabelValue is the name of this operator.
const ManagedByLabelValue = "stateless-pg"

// ManagedByLabelKey is the standard Kubernetes label key for managed-by.
const ManagedByLabelKey = "app.kubernetes.io/managed-by"

// managedByOperatorLabel is the legacy label key for managed-by.
const managedByOperatorLabel = "managed-by"

// ObjectOption is a function that modifies a metav1.Object.
type ObjectOption func(metav1.Object)

// Map is a type alias for string map.
type Map map[string]string

// Merge merges the given map into the current map.
// The given map takes precedence over the current map.
func (m Map) Merge(other map[string]string) Map {
	result := make(Map, len(m)+len(other))
	for k, v := range m {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

// WithLabels merges the given labels with the existing object's labels.
// The given labels take precedence over the existing ones.
func WithLabels(labels map[string]string) ObjectOption {
	return func(o metav1.Object) {
		l := Map{}
		l = l.Merge(labels)
		l = l.Merge(o.GetLabels())

		o.SetLabels(l)
	}
}

// WithAnnotations merges the given annotations with the existing object's annotations.
// The given annotations take precedence over the existing ones.
func WithAnnotations(annotations map[string]string) ObjectOption {
	return func(o metav1.Object) {
		a := Map{}
		a = a.Merge(annotations)
		a = a.Merge(o.GetAnnotations())

		o.SetAnnotations(a)
	}
}

// UpdateObject updates the object's metadata with the provided options. It
// automatically injects the "managed-by" and "app.kubernetes.io/managed-by"
// labels which identifies the operator as the managing entity.
func UpdateObject(o metav1.Object, opts ...ObjectOption) {
	WithLabels(map[string]string{
		managedByOperatorLabel: ManagedByLabelValue,
		ManagedByLabelKey:      ManagedByLabelValue,
	})(o)

	for _, opt := range opts {
		opt(o)
	}
}

type Owner interface {
	metav1.ObjectMetaAccessor
	schema.ObjectKind
}

// WithOwner adds the given object to the list of owner references.
func WithOwner(owner Owner) ObjectOption {
	return func(o metav1.Object) {
		o.SetOwnerReferences(
			append(
				o.GetOwnerReferences(),
				metav1.OwnerReference{
					APIVersion: owner.GroupVersionKind().GroupVersion().String(),
					Kind:       owner.GroupVersionKind().Kind,
					Name:       owner.GetObjectMeta().GetName(),
					UID:        owner.GetObjectMeta().GetUID(),
				},
			),
		)
	}
}
