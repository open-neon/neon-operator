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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NeonClusterKind = "NeonCluster"
	NeonClusterKey  = "neoncluster"
	NeonClusterName = "neonclusters"
)

// NeonClusterSpec defines the desired state of NeonCluster.
// +k8s:openapi-gen=true
type NeonClusterSpec struct {
	// +optional
	SafeKeeper *SafeKeeperSpec `json:"safeKeeper,omitempty"`

	// +optional
	Pageserver *PageServerSpec `json:"pageServer,omitempty"`

	// +optional
	StorageBroker *StorageBrokerSpec `json:"storageBroker,omitempty"`
}

// SafeKeeperSpec defines the desired state of Safekeeper.
// It will be statefullset.
// +k8s:openapi-gen=true
type SafeKeeperSpec struct {
	CommonFields `json:",inline"`

	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=3
	// +optional
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=3
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// storage defines the storage used by Prometheus.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// volumes allows the configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// volumeMounts allows the configuration of additional VolumeMounts.
	//
	// VolumeMounts will be appended to other VolumeMounts in the 'prometheus'
	// container, that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// persistentVolumeClaimRetentionPolicy defines the field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.
	// The default behavior is all PVCs are retained.
	// This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.
	// It requires enabling the StatefulSetAutoDeletePVC feature gate.
	//
	// +optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`
}

// PageServerSpec defines the desired state of PageServerSpec.
// It will be statefullSet.
// +k8s:openapi-gen=true
type PageServerSpec struct {
	CommonFields `json:",inline"`
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// storage defines the storage used by Prometheus.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// volumes allows the configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`
	// volumeMounts allows the configuration of additional VolumeMounts.
	//
	// VolumeMounts will be appended to other VolumeMounts in the 'prometheus'
	// container, that are generated as a result of StorageSpec objects.
	// +optional
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// persistentVolumeClaimRetentionPolicy defines the field controls if and how PVCs are deleted during the lifecycle of a StatefulSet.
	// The default behavior is all PVCs are retained.
	// This is an alpha field from kubernetes 1.23 until 1.26 and a beta field from 1.26.
	// It requires enabling the StatefulSetAutoDeletePVC feature gate.
	//
	// +optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`
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

// NeonClusterStatus defines the observed state of NeonCluster.
type NeonClusterStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the NeonCluster resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="neon-operator",shortName="nc"
// +kubebuilder:printcolumn:name="SafeKeeper Replicas",type="integer",JSONPath=".spec.safeKeeper.minReplicas",description="The minimum number of SafeKeeper replicas"
// +kubebuilder:printcolumn:name="PageServer Replicas",type="integer",JSONPath=".spec.pageServer.minReplicas",description="The minimum number of PageServer replicas"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status

// NeonCluster is the Schema for the neonclusters API
type NeonCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of NeonCluster
	// +required
	Spec NeonClusterSpec `json:"spec"`

	// status defines the observed state of NeonCluster
	// +optional
	Status NeonClusterStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// NeonClusterList contains a list of NeonCluster
// +k8s:openapi-gen=true
type NeonClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []NeonCluster `json:"items"`
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

func init() {
	SchemeBuilder.Register(&NeonCluster{}, &NeonClusterList{})
}
