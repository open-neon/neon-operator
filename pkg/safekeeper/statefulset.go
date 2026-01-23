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
	"github.com/stateless-pg/stateless-pg/pkg/api/v1alpha1"
	"github.com/stateless-pg/stateless-pg/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NeonDefaultImage = "ghcr.io/neondatabase/neon:latest"
)

// makeSafeKeeperStatefulSet creates a StatefulSet for the SafeKeeper component
func makeSafeKeeperStatefulSet(sk *v1alpha1.SafeKeeper, skp *v1alpha1.SafeKeeperProfile, spec *appsv1.StatefulSetSpec) (*appsv1.StatefulSet, error) {

	statefulSet := &appsv1.StatefulSet{
		Spec: *spec,
	}

	operator.UpdateObject(statefulSet,
		operator.WithLabels(sk.Labels),
		operator.WithOwner(sk),
	)

	return statefulSet, nil
}

func makeSafeKeeperStatefulSetSpec(skp *v1alpha1.SafeKeeperProfile) (*appsv1.StatefulSetSpec, error) {
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

	container := corev1.Container{
		Name:            "safekeeper",
		Image:           image,
		ImagePullPolicy: cpf.ImagePullPolicy,
		Resources:       cpf.Resources,
		VolumeMounts:    skp.Spec.VolumeMounts,
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

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
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
