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
	StorageBrokerProfileKind = "StorageBrokerProfile"
	StorageBrokerProfileKey  = "storagebrokerprofile"
	StorageBrokerProfileName = "storagebrokerprofiles"
)

// StorageBrokerConfigOptions defines optional configuration for the StorageBroker
type StorageBrokerConfigOptions struct {
	// timelineChanSize defines the size of the timeline channel buffer
	// +kubebuilder:default=32
	// +kubebuilder:validation:Minimum=1
	// +optional
	TimelineChanSize int32 `json:"timelineChanSize,omitempty"`

	// allKeysChanSize defines the size of the all keys channel buffer
	// +kubebuilder:default=16384
	// +kubebuilder:validation:Minimum=1
	// +optional
	AllKeysChanSize int32 `json:"allKeysChanSize,omitempty"`

	// http2KeepaliveInterval defines the HTTP/2 keepalive interval (e.g., "5000ms", "10s")
	// +kubebuilder:default="5000ms"
	// +optional
	HTTP2KeepaliveInterval string `json:"http2KeepaliveInterval,omitempty"`

	// logFormat defines the log format ("plain" or "json")
	// +kubebuilder:default="plain"
	// +kubebuilder:validation:Enum="plain";"json"
	// +optional
	LogFormat string `json:"logFormat,omitempty"`

	// tlsSecretRef contains cert/key.
	// +optional
	TLSSecretRef *v1.SecretReference `json:"tlsSecretRef,omitempty"`

	// sslCertReloadPeriod defines the SSL certificate reload period (e.g., "60s", "30s")
	// +kubebuilder:default="60s"
	// +optional
	SSLCertReloadPeriod *string `json:"sslCertReloadPeriod,omitempty"`

	// enableTLS indicates whether to enable TLS for StorageBroker
	// +kubebuilder:default=false
	// +optional
	EnableTLS 	   bool `json:"enableTLS,omitempty"`
}

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

	StorageBrokerConfigOptions `json:"inline"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="stateless-pg",shortName="sbp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// StorageBrokerProfile is the Schema for the storagebrokerprofiles API
type StorageBrokerProfile struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of StorageBrokerProfile
	// +required
	Spec StorageBrokerProfileSpec `json:"spec"`
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
