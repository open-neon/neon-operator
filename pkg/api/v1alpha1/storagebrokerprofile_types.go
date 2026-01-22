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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StorageBrokerProfileKind = "StorageBrokerProfile"
	StorageBrokerProfileKey  = "storagebrokerprofile"
	StorageBrokerProfileName = "storagebrokerprofiles"
)

// StorageBrokerProfileSpec defines the desired state of StorageBrokerProfile.
// It will be deployed as a Deployment.
// +k8s:openapi-gen=true
type StorageBrokerProfileSpec struct {
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

// StorageBrokerProfileStatus defines the observed state of StorageBrokerProfile.
type StorageBrokerProfileStatus struct {
	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the StorageBrokerProfile resource.
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
// +kubebuilder:resource:categories="neon-operator",shortName="sbp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status

// StorageBrokerProfile is the Schema for the storagebrokerprofiles API
type StorageBrokerProfile struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of StorageBrokerProfile
	// +required
	Spec StorageBrokerProfileSpec `json:"spec"`

	// status defines the observed state of StorageBrokerProfile
	// +optional
	Status StorageBrokerProfileStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// StorageBrokerProfileList contains a list of StorageBrokerProfile
// +k8s:openapi-gen=true
type StorageBrokerProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []StorageBrokerProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageBrokerProfile{}, &StorageBrokerProfileList{})
}
