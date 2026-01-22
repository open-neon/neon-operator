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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
// It contains TypeMeta and a reduced ObjectMeta.
type EmbeddedPersistentVolumeClaim struct {
	// TypeMeta defines the versioned schema of this representation of an object.
	metav1.TypeMeta `json:",inline"`
	// metadata defines EmbeddedMetadata contains metadata relevant to an EmbeddedResource.
	// +optional
	EmbeddedObjectMetadata `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// spec defines the specification of the  characteristics of a volume requested by a pod author.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec v1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// status is deprecated: this field is never set.
	// +optional
	Status v1.PersistentVolumeClaimStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta
// Only fields which are relevant to embedded resources are included.
type EmbeddedObjectMetadata struct {
	// name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// labels define the map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	// +optional
	//nolint:kubeapilinter
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// annotations defines an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
	// +optional
	//nolint:kubeapilinter
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// StorageSpec defines the configured storage for a group Prometheus servers.
// If no storage option is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.
//
// If multiple storage options are specified, priority will be given as follows:
//  1. emptyDir
//  2. ephemeral
//  3. volumeClaimTemplate
//
// +k8s:openapi-gen=true
type StorageSpec struct {
	// disableMountSubPath deprecated: subPath usage will be removed in a future release.
	// +optional
	DisableMountSubPath bool `json:"disableMountSubPath,omitempty"` // nolint:kubeapilinter
	// emptyDir to be used by the StatefulSet.
	// If specified, it takes precedence over `ephemeral` and `volumeClaimTemplate`.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	// +optional
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// ephemeral to be used by the StatefulSet.
	// This is a beta field in k8s 1.21 and GA in 1.15.
	// For lower versions, starting with k8s 1.19, it requires enabling the GenericEphemeralVolume feature gate.
	// More info: https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#generic-ephemeral-volumes
	// +optional
	Ephemeral *v1.EphemeralVolumeSource `json:"ephemeral,omitempty"`
	// volumeClaimTemplate defines the PVC spec to be used by the Prometheus StatefulSets.
	// The easiest way to use a volume that cannot be automatically provisioned
	// is to use a label selector alongside manually created PersistentVolumes.
	// +optional
	VolumeClaimTemplate EmbeddedPersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// StorageBrokerSpec defines the desired state of StorageBrokerSpec
// It will be deployment.
// +k8s:openapi-gen=true
type StorageBrokerSpec struct {
	CommonFields `json:",inline"`
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`
}

// +k8s:deepcopy-gen=true
type CommonFields struct {
	// kubebuilder:default="ghcr.io/neondatabase/neon:latest"
	// +optional
	Image *string `json:"image,omitempty"`
	// imagePullPolicy defines the image pull policy for the 'prometheus', 'init-config-reloader' and 'config-reloader' containers.
	// See https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy for more details.
	// +kubebuilder:validation:Enum="";Always;Never;IfNotPresent
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// imagePullSecrets defines an optional list of references to Secrets in the same namespace
	// to use for pulling images from registries.
	// See http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// resources defines the resources requests and limits of the 'prometheus' container.
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// nodeSelector defines on which Nodes the Pods are scheduled.
	// +optional
	//nolint:kubeapilinter
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// affinity defines the Pods' affinity scheduling rules if specified.
	// +optional
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// securityContext holds pod-level security attributes and common container settings.
	// This defaults to the default PodSecurityContext.
	// +optional
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
}