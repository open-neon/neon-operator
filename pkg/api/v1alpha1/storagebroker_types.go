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
	StorageBrokerKind = "StorageBroker"
	StorageBrokerKey  = "storagebroker"
	StorageBrokerName = "storagebrokers"
)

// StorageBrokerSpec defines the desired state of StorageBroker.
// +k8s:openapi-gen=true
type StorageBrokerSpec struct {
	// profileRef is a reference to the StorageBrokerProfile resource to use
	// +required
	ProfileRef *v1.ObjectReference `json:"profileRef"`

	// replicas defines the number of StorageBroker replicas to maintain
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// tlsSecretRef contains cert/key for mtls between services.
	// +optional
	TLSSecretRef *v1.SecretReference `json:"tlsSecretRef,omitempty"`
}

// StorageBrokerStatus defines the observed state of StorageBroker.
// +k8s:openapi-gen=true
type StorageBrokerStatus struct {
	// conditions represent the current state of the StorageBroker resource.
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
// +kubebuilder:resource:categories="stateless-pg",shortName="sb"
// +kubebuilder:printcolumn:name="Available",type="string",JSONPath=".status.conditions[?(@.type == 'Available')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status

// StorageBroker is the Schema for the storagebrokers API
type StorageBroker struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of StorageBroker
	// +required
	Spec StorageBrokerSpec `json:"spec"`

	// status defines the observed state of StorageBroker
	// +optional
	Status StorageBrokerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// StorageBrokerList contains a list of StorageBroker
// +k8s:openapi-gen=true
type StorageBrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []StorageBroker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageBroker{}, &StorageBrokerList{})
}
