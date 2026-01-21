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

package neoncluster

import (
	"github.com/open-neon/neon-operator/pkg/api/v1alpha1"
	"github.com/open-neon/neon-operator/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NeonDefaultImage = "ghcr.io/neondatabase/neon:latest"
)

func makePageServerStatefullSet(nc *v1alpha1.NeonCluster) (*appsv1.StatefulSet, error) {
	spec, err := makePageServerStatefullSetSpec(nc)
	if err != nil {
		return nil, err
	}

	statefulSet := &appsv1.StatefulSet{
		Spec: *spec,
	}

	operator.UpdateObject(statefulSet,
		operator.WithLabels(map[string]string{
			"neoncluster": nc.Name,
			"app":         "pageserver",
		}),
	)

	return statefulSet, nil
}

func makePageServerStatefullSetSpec(nc *v1alpha1.NeonCluster) (*appsv1.StatefulSetSpec, error) {
	cpf := nc.Spec.Pageserver.CommonFields
	ps := nc.Spec.Pageserver

	image := NeonDefaultImage
	if cpf.Image != nil {
		image = *cpf.Image
	}

	// Set replicas (using MinReplicas as the desired count)
	replicas := int32(1)
	if ps.MinReplicas != nil {
		replicas = int32(*ps.MinReplicas)
	}

	// Build pod labels
	labels := map[string]string{
		"app":         "pageserver",
		"component":   "pageserver",
		"neoncluster": nc.Name,
	}

	container := corev1.Container{
		Name:            "pageserver",
		Image:           image,
		ImagePullPolicy: cpf.ImagePullPolicy,
		Resources:       cpf.Resources,
		VolumeMounts:    ps.VolumeMounts,
	}

	// Add storage volume mount if storage is specified
	if ps.Storage != nil {
		if ps.Storage.EmptyDir == nil && ps.Storage.Ephemeral == nil {
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
			Volumes:          ps.Volumes,
		},
	}

	// Add storage volumes if specified
	if ps.Storage != nil {
		if ps.Storage.EmptyDir != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: ps.Storage.EmptyDir,
				},
			})
		} else if ps.Storage.Ephemeral != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					Ephemeral: ps.Storage.Ephemeral,
				},
			})
		}
	}

	spec := appsv1.StatefulSetSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		ServiceName:                          "pageserver",
		Template:                             podTemplateSpec,
		PersistentVolumeClaimRetentionPolicy: ps.PersistentVolumeClaimRetentionPolicy,
	}

	// Add VolumeClaimTemplates if persistent storage is configured
	if ps.Storage != nil && ps.Storage.EmptyDir == nil && ps.Storage.Ephemeral == nil {
		pvc := ps.Storage.VolumeClaimTemplate
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
