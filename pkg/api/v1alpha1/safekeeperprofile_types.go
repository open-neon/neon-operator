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

// SafeKeeperConfigOptions defines all command-line arguments for the safekeeper process
// +k8s:openapi-gen=true
type SafeKeeperConfigOptions struct {

	// listenPgTenantOnly specifies the tenant-scoped WAL service endpoint
	// +optional
	ListenPgTenantOnly *string `json:"listenPgTenantOnly,omitempty"`

	// Node & Cluster Configuration
	// ============================

	// datadir is the path to the safekeeper data directory
	// +kubebuilder:default="./data"
	// +optional
	DataDir string `json:"dataDir,omitempty"`

	// availabilityZone is an identifier for the safekeeper's availability zone
	// +optional
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// brokerKeepaliveInterval specifies the keepalive ping interval (e.g., 15s, 5m)
	// +kubebuilder:default="15s"
	// +optional
	BrokerKeepaliveInterval string `json:"brokerKeepaliveInterval,omitempty"`

	// heartbeatTimeout specifies the peer safekeeper heartbeat timeout
	// +kubebuilder:default="5s"
	// +optional
	HeartbeatTimeout string `json:"heartbeatTimeout,omitempty"`

	// peerRecovery enables/disables peer recovery
	// +kubebuilder:default=false
	// +optional
	PeerRecovery bool `json:"peerRecovery,omitempty"`

	// WAL Management & Storage
	// =======================

	// maxOffloaderLag specifies max LAG before safekeeper elected for offloading (in bytes)
	// +kubebuilder:default=134217728
	// +optional
	MaxOffloaderLag int64 `json:"maxOffloaderLag,omitempty"`

	// maxReelectOffloaderLagBytes triggers re-election if offloader lags by this amount
	// +kubebuilder:default=0
	// +optional
	MaxReelectOffloaderLagBytes int64 `json:"maxReelectOffloaderLagBytes,omitempty"`

	// maxTimelineDiskUsageBytes specifies max WAL disk per timeline (0 = disabled)
	// +kubebuilder:default=0
	// +optional
	MaxTimelineDiskUsageBytes int64 `json:"maxTimelineDiskUsageBytes,omitempty"`

	// walBackupParallelJobs specifies max parallel WAL segment uploads
	// +kubebuilder:default=5
	// +optional
	WalBackupParallelJobs int64 `json:"walBackupParallelJobs,omitempty"`

	// disableWalBackup disables WAL backup to remote storage
	// +kubebuilder:default=false
	// +optional
	DisableWalBackup bool `json:"disableWalBackup,omitempty"`

	// remoteStorageMaxConcurrentSyncs specifies max concurrent syncs to remote storage
	// +optional
	RemoteStorageMaxConcurrentSyncs *int64 `json:"remoteStorageMaxConcurrentSyncs,omitempty"`

	// remoteStorageMaxSyncErrors specifies max sync errors before failure
	// +optional
	RemoteStorageMaxSyncErrors *int64 `json:"remoteStorageMaxSyncErrors,omitempty"`

	// Authentication & Security
	// ========================

	// pgAuthPublicKeyPath specifies the JWT public key for WAL service auth
	// +optional
	PgAuthPublicKeyPath *string `json:"pgAuthPublicKeyPath,omitempty"`

	// pgTenantOnlyAuthPublicKeyPath specifies the JWT public key for tenant-only WAL auth
	// +optional
	PgTenantOnlyAuthPublicKeyPath *string `json:"pgTenantOnlyAuthPublicKeyPath,omitempty"`

	// httpAuthPublicKeyPath specifies the JWT public key for HTTP service auth
	// +optional
	HttpAuthPublicKeyPath *string `json:"httpAuthPublicKeyPath,omitempty"`

	// authTokenPath path to JWT token file for peer authentication
	// +optional
	AuthTokenPath *string `json:"authTokenPath,omitempty"`

	// sslKeyFile path to HTTPS private key file
	// +kubebuilder:default="server.key"
	// +optional
	SslKeyFile *string `json:"sslKeyFile,omitempty"`

	// sslCertFile path to HTTPS certificate file
	// +kubebuilder:default="server.crt"
	// +optional
	SslCertFile *string `json:"sslCertFile,omitempty"`

	// sslCertReloadPeriod certificate reload interval
	// +kubebuilder:default="60s"
	// +optional
	SslCertReloadPeriod *string `json:"sslCertReloadPeriod,omitempty"`

	// sslCaFile path to trusted CA certificate file
	// +optional
	SslCaFile *string `json:"sslCaFile,omitempty"`

	// useHttpsSafekeeperApi uses HTTPS for peer safekeeper API
	// +kubebuilder:default=false
	// +optional
	UseHttpsSafekeeperApi bool `json:"useHttpsSafekeeperApi,omitempty"`

	// enableTlsWalServiceApi enables TLS in WAL service API
	// +kubebuilder:default=false
	// +optional
	EnableTlsWalServiceApi bool `json:"enableTlsWalServiceApi,omitempty"`

	// Safety & Reliability
	// ===================

	// noSync disables fsync (unsafe, for testing only)
	// +kubebuilder:default=false
	// +optional
	NoSync bool `json:"noSync,omitempty"`

	// peerRecovery enables/disables peer recovery
	// +kubebuilder:default=false
	// +optional
	PeerRecoveryEnabled bool `json:"peerRecoveryEnabled,omitempty"`

	// enableOffload enables automatic switching to offloaded state
	// +kubebuilder:default=false
	// +optional
	EnableOffload bool `json:"enableOffload,omitempty"`

	// deleteOffloadedWal deletes local WAL after offloading
	// +kubebuilder:default=false
	// +optional
	DeleteOffloadedWal bool `json:"deleteOffloadedWal,omitempty"`

	// Backup & Maintenance
	// ===================

	// partialBackupTimeout wait time before uploading partial segment
	// +kubebuilder:default="15m"
	// +optional
	PartialBackupTimeout string `json:"partialBackupTimeout,omitempty"`

	// partialBackupConcurrency concurrent partial segment uploads
	// +kubebuilder:default=5
	// +optional
	PartialBackupConcurrency int64 `json:"partialBackupConcurrency,omitempty"`

	// controlFileSaveInterval auto-save control file interval
	// +kubebuilder:default="300s"
	// +optional
	ControlFileSaveInterval string `json:"controlFileSaveInterval,omitempty"`

	// evictionMinResident minimum timeline residency before eviction
	// +kubebuilder:default="15m"
	// +optional
	EvictionMinResident string `json:"evictionMinResident,omitempty"`

	// disablePeriodicBrokerPush disables broker push (testing only)
	// +kubebuilder:default=false
	// +optional
	DisablePeriodicBrokerPush bool `json:"disablePeriodicBrokerPush,omitempty"`

	// Performance & Monitoring
	// ======================

	// logFormat specifies the logging format (plain or json)
	// +kubebuilder:default="plain"
	// +kubebuilder:validation:Enum=plain;json
	// +optional
	LogFormat string `json:"logFormat,omitempty"`

	// forceMetricCollectionOnScrape collects metrics on each scrape
	// +kubebuilder:default=true
	// +optional
	ForceMetricCollectionOnScrape bool `json:"forceMetricCollectionOnScrape,omitempty"`

	// Concurrency & Sharding
	// =====================

	// walReaderFanout enables fanning out WAL to different shards
	// +kubebuilder:default=false
	// +optional
	WalReaderFanout bool `json:"walReaderFanout,omitempty"`

	// maxDeltaForFanout maximum position delta for fanout
	// +optional
	MaxDeltaForFanout *int64 `json:"maxDeltaForFanout,omitempty"`

	// currentThreadRuntime runs in single-threaded mode (debugging only)
	// +kubebuilder:default=false
	// +optional
	CurrentThreadRuntime bool `json:"currentThreadRuntime,omitempty"`

	// walsendersKeepHorizon keeps WAL for replication connections
	// +kubebuilder:default=false
	// +optional
	WalsendersKeepHorizon bool `json:"walsendersKeepHorizon,omitempty"`

	// Disk Management
	// ==============

	// globalDiskCheckInterval disk usage check interval
	// +kubebuilder:default="60s"
	// +optional
	GlobalDiskCheckInterval string `json:"globalDiskCheckInterval,omitempty"`

	// maxGlobalDiskUsageRatio portion of filesystem capacity for all timelines (0.0 = disabled)
	// +kubebuilder:default=0.0
	// +optional
	MaxGlobalDiskUsageRatio float64 `json:"maxGlobalDiskUsageRatio,omitempty"`

	// Development & Debugging
	// =====================

	// dev enables development mode (disables security checks)
	// +kubebuilder:default=false
	// +optional
	Dev bool `json:"dev,omitempty"`

	// enablePullTimelineOnStartup auto-pulls timelines from peer safekeepers
	// +kubebuilder:default=false
	// +optional
	EnablePullTimelineOnStartup bool `json:"enablePullTimelineOnStartup,omitempty"`
}

const (
	SafeKeeperProfileKind = "SafeKeeperProfile"
	SafeKeeperProfileKey  = "safekeeperprofile"
	SafeKeeperProfileName = "safekeeperprofiles"
)

// SafeKeeperProfileSpec defines the desired state of SafeKeeperProfile.
// +k8s:openapi-gen=true
type SafeKeeperProfileSpec struct {
	CommonFields `json:",inline"`

	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=3
	// +optional
	MinReplicas *int64 `json:"minReplicas,omitempty"`

	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=3
	// +optional
	MaxReplicas *int64 `json:"maxReplicas,omitempty"`

	// storage defines the storage used by SafeKeeper.
	// +optional
	Storage *StorageSpec `json:"storage,omitempty"`

	// volumes allows the configuration of additional volumes on the output
	// StatefulSet definition. Volumes specified will be appended to other
	// volumes that are generated as a result of StorageSpec objects.
	// +optional
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// volumeMounts allows the configuration of additional VolumeMounts.
	//
	// VolumeMounts will be appended to other VolumeMounts in the 'safekeeper'
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

	// SafeKeeperCliArgs defines all safekeeper command-line arguments
	SafeKeeperConfigOptions `json:",inline"`
}

// +genclient
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories="stateless-pg",shortName="skp"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// SafeKeeperProfile is the Schema for the safekeeperprofiles API
type SafeKeeperProfile struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of SafeKeeperProfile
	// +required
	Spec SafeKeeperProfileSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// SafeKeeperProfileList contains a list of SafeKeeperProfile
// +k8s:openapi-gen=true
type SafeKeeperProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []SafeKeeperProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SafeKeeperProfile{}, &SafeKeeperProfileList{})
}
