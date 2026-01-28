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

package safekeeper

import (
	"fmt"
	"strings"

	"github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	"github.com/stateless-pg/stateless-pg/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NeonDefaultImage = "ghcr.io/neondatabase/neon:latest"
)

// buildSafeKeeperArgs builds the command-line arguments for the safekeeper process
func buildSafeKeeperArgs(nodeID int32, sf *v1alpha1.SafeKeeper, opts *v1alpha1.SafeKeeperConfigOptions) []string {
	args := []string{
		fmt.Sprintf("--id=%d", nodeID),
	}

	if opts == nil {
		return args
	}

	args = append(args, fmt.Sprintf("--listen_pg=%s", "0.0.0.0:5432"))
	args = append(args, fmt.Sprintf("--listen_http=%s", "0.0.0.0:9898"))
	args = append(args, fmt.Sprintf("--broker_endpoint=http://%s-broker:50051", sf.Labels["neoncluster"]))

	// Node & Cluster Configuration
	args = append(args, fmt.Sprintf("--datadir=%s", opts.DataDir))

	if opts.AvailabilityZone != nil {
		args = append(args, fmt.Sprintf("--availability_zone=%s", *opts.AvailabilityZone))
	}

	args = append(args, fmt.Sprintf("--broker_keepalive_interval=%s", opts.BrokerKeepaliveInterval))

	args = append(args, fmt.Sprintf("--heartbeat_timeout=%s", opts.HeartbeatTimeout))

	if opts.PeerRecovery {
		args = append(args, "--peer_recovery=true")
	}

	args = append(args, fmt.Sprintf("--max_offloader_lag=%d", opts.MaxOffloaderLag))

	args = append(args, fmt.Sprintf("--max_reelect_offloader_lag_bytes=%d", opts.MaxReelectOffloaderLagBytes))

	args = append(args, fmt.Sprintf("--max_timeline_disk_usage_bytes=%d", opts.MaxTimelineDiskUsageBytes))

	args = append(args, fmt.Sprintf("--wal_backup_parallel_jobs=%d", opts.WalBackupParallelJobs))

	if opts.DisableWalBackup {
		args = append(args, "--disable_wal_backup=true")
	}

	// Remote Storage Configuration
	if opts.RemoteStorageMaxConcurrentSyncs != nil || opts.RemoteStorageMaxSyncErrors != nil ||
		opts.RemoteStorageBucketName != nil || opts.RemoteStorageBucketRegion != nil ||
		opts.RemoteStorageConcurrencyLimit != nil {
		var remoteStorageParts []string
		remoteStorageParts = append(remoteStorageParts, "{")
		if opts.RemoteStorageMaxConcurrentSyncs != nil {
			remoteStorageParts = append(remoteStorageParts, fmt.Sprintf("max_concurrent_syncs = %d", *opts.RemoteStorageMaxConcurrentSyncs))
		}
		if opts.RemoteStorageMaxSyncErrors != nil {
			remoteStorageParts = append(remoteStorageParts, fmt.Sprintf("max_sync_errors = %d", *opts.RemoteStorageMaxSyncErrors))
		}
		if opts.RemoteStorageBucketName != nil && *opts.RemoteStorageBucketName != "" {
			remoteStorageParts = append(remoteStorageParts, fmt.Sprintf("bucket_name = \"%s\"", *opts.RemoteStorageBucketName))
		}
		if opts.RemoteStorageBucketRegion != nil && *opts.RemoteStorageBucketRegion != "" {
			remoteStorageParts = append(remoteStorageParts, fmt.Sprintf("bucket_region = \"%s\"", *opts.RemoteStorageBucketRegion))
		}
		if opts.RemoteStorageConcurrencyLimit != nil {
			remoteStorageParts = append(remoteStorageParts, fmt.Sprintf("concurrency_limit = %d", *opts.RemoteStorageConcurrencyLimit))
		}
		remoteStorageParts = append(remoteStorageParts, "}")
		remoteStorageStr := strings.Join(remoteStorageParts, " ")
		remoteStorageStr = strings.TrimSuffix(strings.TrimSuffix(remoteStorageStr, ", }"), ",} ")
		remoteStorageStr = strings.Replace(remoteStorageStr, ", }", " }", 1)
		args = append(args, fmt.Sprintf("--remote_storage=%s", remoteStorageStr))
	}

	// Authentication & Security
	if opts.PgAuthPublicKeyPath != nil {
		args = append(args, fmt.Sprintf("--pg_auth_public_key_path=%s", *opts.PgAuthPublicKeyPath))
	}
	if opts.PgTenantOnlyAuthPublicKeyPath != nil {
		args = append(args, fmt.Sprintf("--pg_tenant_only_auth_public_key_path=%s", *opts.PgTenantOnlyAuthPublicKeyPath))
	}
	if opts.HttpAuthPublicKeyPath != nil {
		args = append(args, fmt.Sprintf("--http_auth_public_key_path=%s", *opts.HttpAuthPublicKeyPath))
	}
	if opts.AuthTokenPath != nil {
		args = append(args, fmt.Sprintf("--auth_token_path=%s", *opts.AuthTokenPath))
	}
	if opts.SslKeyFile != nil {
		args = append(args, fmt.Sprintf("--ssl_key_file=%s", *opts.SslKeyFile))
	}
	if opts.SslCertFile != nil {
		args = append(args, fmt.Sprintf("--ssl_cert_file=%s", *opts.SslCertFile))
	}
	if opts.SslCertReloadPeriod != nil {
		args = append(args, fmt.Sprintf("--ssl_cert_reload_period=%s", *opts.SslCertReloadPeriod))
	}
	if opts.SslCaFile != nil {
		args = append(args, fmt.Sprintf("--ssl_ca_file=%s", *opts.SslCaFile))
	}
	if opts.UseHttpsSafekeeperApi {
		args = append(args, "--use_https_safekeeper_api=true")
	}
	if opts.EnableTlsWalServiceApi {
		args = append(args, "--enable_tls_wal_service_api=true")
	}

	// Safety & Reliability
	if opts.NoSync {
		args = append(args, "--no_sync=true")
	}
	if opts.PeerRecoveryEnabled {
		args = append(args, "--peer_recovery=true")
	}
	if opts.EnableOffload {
		args = append(args, "--enable_offload=true")
	}
	if opts.DeleteOffloadedWal {
		args = append(args, "--delete_offloaded_wal=true")
	}

	args = append(args, fmt.Sprintf("--partial_backup_timeout=%s", opts.PartialBackupTimeout))

	args = append(args, fmt.Sprintf("--partial_backup_concurrency=%d", opts.PartialBackupConcurrency))

	args = append(args, fmt.Sprintf("--control_file_save_interval=%s", opts.ControlFileSaveInterval))

	args = append(args, fmt.Sprintf("--eviction_min_resident=%s", opts.EvictionMinResident))

	if opts.DisablePeriodicBrokerPush {
		args = append(args, "--disable_periodic_broker_push=true")
	}

	// Performance & Monitoring

	args = append(args, fmt.Sprintf("--log_format=%s", opts.LogFormat))

	if !opts.ForceMetricCollectionOnScrape {
		args = append(args, "--force_metric_collection_on_scrape=false")
	}

	// Concurrency & Sharding
	if opts.WalReaderFanout {
		args = append(args, "--wal_reader_fanout=true")
	}
	if opts.MaxDeltaForFanout != nil {
		args = append(args, fmt.Sprintf("--max_delta_for_fanout=%d", *opts.MaxDeltaForFanout))
	}
	if opts.CurrentThreadRuntime {
		args = append(args, "--current_thread_runtime=true")
	}
	if opts.WalsendersKeepHorizon {
		args = append(args, "--walsenders_keep_horizon=true")
	}

	// Disk Management

	args = append(args, fmt.Sprintf("--global_disk_check_interval=%s", opts.GlobalDiskCheckInterval))

	args = append(args, fmt.Sprintf("--max_global_disk_usage_ratio=%f", opts.MaxGlobalDiskUsageRatio))

	// Development & Debugging
	if opts.Dev {
		args = append(args, "--dev", "true")
	}
	if opts.EnablePullTimelineOnStartup {
		args = append(args, "--enable_pull_timeline_on_startup", "true")
	}

	// Additional listen endpoints
	if opts.ListenPgTenantOnly != nil {
		args = append(args, "--listen_pg_tenant_only", *opts.ListenPgTenantOnly)
	}
	if opts.AdvertisePg != nil {
		args = append(args, "--advertise_pg", *opts.AdvertisePg)
	}

	return args
}

// makeSafeKeeperStatefulSet creates a StatefulSet for the SafeKeeper component
func makeSafeKeeperStatefulSet(sk *v1alpha1.SafeKeeper,  spec *appsv1.StatefulSetSpec) (*appsv1.StatefulSet, error) {

	statefulSet := &appsv1.StatefulSet{
		Spec: *spec,
	}

	operator.UpdateObject(statefulSet,
		operator.WithLabels(sk.Labels),
		operator.WithOwner(sk),
	)

	return statefulSet, nil
}

func makeSafeKeeperStatefulSetSpec(sk *v1alpha1.SafeKeeper, skp *v1alpha1.SafeKeeperProfile) (*appsv1.StatefulSetSpec, error) {
	cpf := skp.Spec.CommonFields

	image := NeonDefaultImage
	if cpf.Image != nil {
		image = *cpf.Image
	}

	// Set replicas (using MinReplicas as the desired count)
	replicas := int32(1)
	if skp.Spec.MinReplicas != nil {
		replicas = int32(*skp.Spec.MinReplicas)
	}

	// Build pod labels
	labels := map[string]string{
		"app":       "safekeeper",
		"component": "safekeeper-statefulset",
	}

	// Environment variables for pod identity and configuration
	env := []corev1.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.hostname",
				},
			},
		},
	}

	// Note: For StatefulSet, each pod's nodeID is derived from its ordinal
	// (e.g., pod name "safekeeper-0" gets ID 0, "safekeeper-1" gets ID 1)
	// We use ordinal 0 as a base for the args spec; individual pods
	// must override the --id argument with their actual ordinal
	// This is typically handled by a wrapper script or init container
	args := buildSafeKeeperArgs(0, sk, &skp.Spec.SafeKeeperConfigOptions)

	container := corev1.Container{
		Name:            "safekeeper",
		Image:           image,
		ImagePullPolicy: cpf.ImagePullPolicy,
		Resources:       cpf.Resources,
		VolumeMounts:    skp.Spec.VolumeMounts,
		Args:            args,
		Env:             env,
	}

	// Add storage volume mount if storage is specified
	if skp.Spec.Storage != nil {
		if skp.Spec.Storage.EmptyDir == nil && skp.Spec.Storage.Ephemeral == nil {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      "data",
				MountPath: "/data",
			})
		}
	}

	// Init container to extract pod ordinal and configure node ID
	initContainers := []corev1.Container{
		{
			Name:  "safekeeper-init",
			Image: image,
			Command: []string{
				"sh",
				"-c",
				`# Extract ordinal from pod name (format: <name>-<ordinal>)
POD_NAME=$(hostname)
ORDINAL=${POD_NAME##*-}

# Validate that ordinal is a number
if ! echo "$ORDINAL" | grep -qE '^[0-9]+$'; then
  echo "Failed to extract ordinal from pod name: $POD_NAME"
  exit 1
fi

echo "Pod: $POD_NAME, Ordinal: $ORDINAL" >&2`,
			},
			Env: []corev1.EnvVar{
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
			},
		},
	}

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			InitContainers:   initContainers,
			Containers:       []corev1.Container{container},
			ImagePullSecrets: cpf.ImagePullSecrets,
			NodeSelector:     cpf.NodeSelector,
			Affinity:         cpf.Affinity,
			SecurityContext:  cpf.SecurityContext,
			Volumes:          skp.Spec.Volumes,
		},
	}

	// Add storage volumes if specified
	if skp.Spec.Storage != nil {
		if skp.Spec.Storage.EmptyDir != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: skp.Spec.Storage.EmptyDir,
				},
			})
		} else if skp.Spec.Storage.Ephemeral != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					Ephemeral: skp.Spec.Storage.Ephemeral,
				},
			})
		}
	}

	spec := appsv1.StatefulSetSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		ServiceName:                          "safekeeper",
		Template:                             podTemplateSpec,
		PersistentVolumeClaimRetentionPolicy: skp.Spec.PersistentVolumeClaimRetentionPolicy,
	}

	// Add VolumeClaimTemplates if persistent storage is configured
	if skp.Spec.Storage != nil && skp.Spec.Storage.EmptyDir == nil && skp.Spec.Storage.Ephemeral == nil {
		pvc := skp.Spec.Storage.VolumeClaimTemplate
		if pvc.EmbeddedObjectMetadata.Name == "" {
			pvc.EmbeddedObjectMetadata.Name = "data"
		}

		spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        pvc.EmbeddedObjectMetadata.Name,
					Labels:      pvc.EmbeddedObjectMetadata.Labels,
					Annotations: pvc.EmbeddedObjectMetadata.Annotations,
				},
				Spec: pvc.Spec,
			},
		}
	}

	return &spec, nil
}
