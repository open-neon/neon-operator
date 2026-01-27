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
	PageServerKind = "PageServer"
	PageServerKey  = "pageserver"
	PageServerName = "pageservers"
)

// PageServerSpec defines the desired state of PageServer.
// +k8s:openapi-gen=true
type PageServerSpec struct {
	// profileRef is a reference to the PageServerProfile resource to use
	// +required
	ProfileRef *v1.ObjectReference `json:"profileRef"`

	// replicas defines the number of PageServer replicas to maintain
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// nodeId is the unique identifier for this PageServer instance
	// +optional
	NodeId *int32 `json:"nodeId,omitempty"`

	// objectStorage defines the configuration for object storage used by Neon components
	// +required
	ObjectStorage ObjectStorageSpec `json:"objectStorage"`
}

// PageServerStatus defines the observed state of PageServer.
// +k8s:openapi-gen=true
type PageServerStatus struct {
	// conditions represent the current state of the PageServer resource.
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
// +kubebuilder:resource:categories="stateless-pg",shortName="ps"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status

// PageServer is the Schema for the pageservers API
type PageServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PageServer
	// +required
	Spec PageServerSpec `json:"spec"`

	// status defines the observed state of PageServer
	// +optional
	Status PageServerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PageServerList contains a list of PageServer
// +k8s:openapi-gen=true
type PageServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PageServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PageServer{}, &PageServerList{})
}
