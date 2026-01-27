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

const (
	NeonClusterKind = "NeonCluster"
	NeonClusterKey  = "neoncluster"
	NeonClusterName = "neonclusters"
)

// NeonClusterSpec defines the desired state of NeonCluster.
// +k8s:openapi-gen=true
type NeonClusterSpec struct {
	// regionName is the name of the region where the NeonCluster is deployed
	// +optional
	RegionName string `json:"regionName,omitempty"`
	// safekeeperProfileRef is a reference to the SafeKeeperProfile resource
	// +optional
	SafeKeeperProfileRef *v1.ObjectReference `json:"safekeeperProfileRef,omitempty"`

	// pageserverProfileRef is a reference to the PageServerProfile resource
	// +optional
	PageServerProfileRef *v1.ObjectReference `json:"pageserverProfileRef,omitempty"`

	// storageBrokerProfileRef is a reference to the StorageBrokerProfile resource
	// +optional
	StorageBrokerProfileRef *v1.ObjectReference `json:"storageBrokerProfileRef,omitempty"`

	// objectStorage defines the configuration for object storage used by Neon components
	// +required
	ObjectStorage ObjectStorageSpec `json:"objectStorage"`
}

// ObjectStorageSpec defines the configuration for object storage used by Neon components.
// +k8s:openapi-gen=true
type ObjectStorageSpec struct {
	// provider defines the backend.
	// +kubebuilder:validation:Enum=s3;gcs;azure;minio;local
	Provider string `json:"provider"`
	
	// endpoint is the URL of the object storage service
	// +required
	Endpoint string `json:"endpoint"`

	// bucket is the name of the storage bucket
	// +required
	Bucket string `json:"bucket"`

	// region specifies the storage region
	// +required
	Region string `json:"region"`

	// credentialsSecret is a reference to a secret containing object storage credentials
	// +optional
	CredentialsSecret *v1.SecretReference `json:"credentialsSecret,omitempty"`

	// prefix is the path prefix for all objects stored
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// maxConcurrentRequests defines the maximum number of concurrent requests to object storage
	// +optional
	MaxConcurrentRequests *int32 `json:"maxConcurrentRequests,omitempty"`

	// extraConfig allows specifying additional configuration parameters as key-value pairs
	// +optional
	ExtraConfig map[string]string `json:"extraConfig,omitempty"`
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
// +kubebuilder:resource:categories="stateless-pg",shortName="nc"
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

func init() {
	SchemeBuilder.Register(&NeonCluster{}, &NeonClusterList{})
}
