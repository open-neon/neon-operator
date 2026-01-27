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
	PageServerProfileKind = "PageServerProfile"
	PageServerProfileKey  = "pageserverprofile"
	PageServerProfileName = "pageserverprofiles"
)

// PageServerProfileSpec defines the desired state of PageServerProfile.
// +k8s:openapi-gen=true
type PageServerProfileSpec struct {
	// mode defines whether PageServer runs standalone or with a control plane.
	// +kubebuilder:validation:Enum=standalone;managed
	// +kubebuilder:default=managed
	Mode string `json:"mode,omitempty"`

	// controlPlane configures controller connectivity.
	// +optional
	ControlPlane *ControlPlaneSpec `json:"controlPlane,omitempty"`

	// durability controls checkpointing & WAL safety.
	// +optional
	Durability *DurabilitySpec `json:"durability,omitempty"`

	// retention controls GC, PITR, and history.
	// +optional
	Retention *RetentionSpec `json:"retention,omitempty"`

	// performance controls IO & ingestion tuning.
	// +optional
	Performance *PerformanceSpec `json:"performance,omitempty"`

	// security controls auth and TLS.
	// +optional
	Security *SecuritySpec `json:"security,omitempty"`

	// observability controls logs, metrics, tracing.
	// +optional
	Observability *ObservabilitySpec `json:"observability,omitempty"`

	// advanced allows expert-only overrides.
	// +optional
	Advanced *AdvancedSpec `json:"advanced,omitempty"`

	CommonFields `json:",inline"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// storage defines the storage used by PageServer.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// volumes allows the configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// volumeMounts allows the configuration of additional VolumeMounts.
	//
	// VolumeMounts will be appended to other VolumeMounts in the 'pageserver'
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

type ControlPlaneSpec struct {
	// endpoint is the control plane API URL.
	Endpoint string `json:"endpoint"`

	// emergencyMode disables controller dependency (dev/CI only).
	// +kubebuilder:default=false
	EmergencyMode bool `json:"emergencyMode,omitempty"`
}

type DurabilitySpec struct {
	// checkpointDistance bounds WAL before flush.
	// +kubebuilder:default="256Mi"
	CheckpointDistance string `json:"checkpointDistance,omitempty"`

	// checkpointTimeout ensures eventual upload.
	// +kubebuilder:default="10m"
	CheckpointTimeout string `json:"checkpointTimeout,omitempty"`
}

type RetentionSpec struct {
	// historyRetention controls GC horizon.
	// +kubebuilder:default="7d"
	HistoryRetention string `json:"historyRetention,omitempty"`

	// pitrRetention controls PITR branching.
	// +kubebuilder:default="7d"
	PITRRetention string `json:"pitrRetention,omitempty"`

	// gcInterval defines GC frequency.
	// +kubebuilder:default="1h"
	GCInterval string `json:"gcInterval,omitempty"`
}

type PerformanceSpec struct {
	// ioMode controls disk IO behavior.
	// +kubebuilder:validation:Enum=buffered;direct
	// +kubebuilder:default=direct
	IOMode string `json:"ioMode,omitempty"`

	// ingestBatchSize limits WAL ingestion batch.
	// +optional
	IngestBatchSize *int64 `json:"ingestBatchSize,omitempty"`
}

type SecuritySpec struct {
	// enableTLS enables TLS for PageServer APIs.
	// +kubebuilder:default=true
	EnableTLS bool `json:"enableTLS,omitempty"`

	// tlsSecretRef contains cert/key.
	// +optional
	TLSSecretRef *v1.SecretReference `json:"tlsSecretRef,omitempty"`

	// authType controls API auth.
	// +kubebuilder:validation:Enum=none;jwt
	// +kubebuilder:default=jwt
	AuthType string `json:"authType,omitempty"`
}

type ObservabilitySpec struct {
	// logLevel controls verbosity.
	// +kubebuilder:validation:Enum=debug;info;warn;error
	// +kubebuilder:default=info
	LogLevel string `json:"logLevel,omitempty"`

	// metrics enables Prometheus metrics.
	// +kubebuilder:default=true
	Metrics bool `json:"metrics,omitempty"`
}

type AdvancedSpec struct {
	// extraConfig injects raw TOML (expert use only).
	// +optional
	ExtraConfig string `json:"extraConfig,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="stateless-pg",shortName="psp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PageServerProfile is the Schema for the pageserverprofiles API
type PageServerProfile struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PageServerProfile
	// +required
	Spec PageServerProfileSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// PageServerProfileList contains a list of PageServerProfile
// +k8s:openapi-gen=true
type PageServerProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PageServerProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PageServerProfile{}, &PageServerProfileList{})
}
