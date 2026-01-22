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
	"github.com/stateless-pg.io/neon-operator/pkg/api/v1alpha1"
	"github.com/stateless-pg.io/neon-operator/pkg/operator"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// makeSafekeeperDeployment creates a Deployment for the Safekeeper component
// based on the provided NeonCluster specification.
func makeSafekeeperDeployment(nc *v1alpha1.NeonCluster) (*appsv1.Deployment, error) {
	spec, err := makeSafekeeperDeploymentSpec(nc)
	if err != nil {
		return nil, err
	}

	deployment := &appsv1.Deployment{
		Spec: *spec,
	}

	operator.UpdateObject(deployment,
		operator.WithLabels(map[string]string{
			"neoncluster": nc.Name,
			"app":         "safekeeper",
		}),
		operator.WithOwner(nc),
	)

	return deployment, nil
}

func makeSafekeeperDeploymentSpec(nc *v1alpha1.NeonCluster) (*appsv1.DeploymentSpec, error) {
	cpf := nc.Spec.SafeKeeper.CommonFields
	sk := nc.Spec.SafeKeeper

	image := NeonDefaultImage
	if cpf.Image != nil {
		image = *cpf.Image
	}

	// Set replicas (using MinReplicas as the desired count)
	replicas := int32(3)
	if sk.MinReplicas != nil {
		replicas = int32(*sk.MinReplicas)
	}

	// Build pod labels
	labels := map[string]string{
		"app":         "safekeeper",
		"component":   "safekeeper",
		"neoncluster": nc.Name,
	}

	container := corev1.Container{
		Name:            "safekeeper",
		Image:           image,
		ImagePullPolicy: cpf.ImagePullPolicy,
		Resources:       cpf.Resources,
		VolumeMounts:    sk.VolumeMounts,
	}

	// Add storage volume mount if storage is specified
	if sk.Storage != nil {
		if sk.Storage.EmptyDir != nil || sk.Storage.Ephemeral != nil {
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
			Volumes:          sk.Volumes,
		},
	}

	// Add storage volumes if specified (Deployment only supports EmptyDir or Ephemeral)
	if sk.Storage != nil {
		if sk.Storage.EmptyDir != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: sk.Storage.EmptyDir,
				},
			})
		} else if sk.Storage.Ephemeral != nil {
			podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					Ephemeral: sk.Storage.Ephemeral,
				},
			})
		}
	}

	spec := appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: podTemplateSpec,
	}

	return &spec, nil
}
