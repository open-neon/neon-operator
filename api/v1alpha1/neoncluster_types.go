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


// NeonClusterSpec defines the desired state of NeonCluster.
type NeonClusterSpec struct {
	// +optional
	SafeKeeper SafeKeeperSpec `json:"safeKeeper,omitempty"`
	
	// +optional
	Pageserver PageServerSpec `json:"pageServer,omitempty"`

	// +optional
	StorageBroker StorageBrokerSpec `json:"storageBroker,omitempty"`
}

// SafeKeeperSpec defines the desired state of Safekeeper.
// It will be statefullset.
type SafeKeeperSpec struct {
	// +kubebuilder:default=1                                                                                                                                      
    // +kubebuilder:validation:Minimum=1                                                                                                                           
    // +optional                                                                                                                                                   
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// +optional
	Template v1.PodTemplateSpec `json:"template,omitempty"`
}

// PageServerSpec defines the desired state of PageServerSpec.
// It will be statefullSet.
type PageServerSpec struct {
	// +kubebuilder:default=1                                                                                                                                      
    // +kubebuilder:validation:Minimum=1                                                                                                                           
    // +optional                                                                                                                                                   
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// +optional
	Template v1.PodTemplateSpec `json:"template,omitempty"`
}

// StorageBrokerSpec defines the desired state of StorageBrokerSpec
// It will be deployment.
type StorageBrokerSpec struct {
	// +kubebuilder:default=1                                                                                                                                      
    // +kubebuilder:validation:Minimum=1                                                                                                                           
    // +optional                                                                                                                                                   
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// +optional
	Template v1.PodTemplateSpec `json:"template,omitempty"`
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

// +kubebuilder:object:root=true
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
type NeonClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []NeonCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NeonCluster{}, &NeonClusterList{})
}
