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

package storagebroker

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

// makeStorageBrokerDeployment creates a Deployment for the StorageBroker component
func makeStorageBrokerDeployment(sb *v1alpha1.StorageBroker, sbp *v1alpha1.StorageBrokerProfile, spec *appsv1.DeploymentSpec) (*appsv1.Deployment, error) {

	deployment := &appsv1.Deployment{
		Spec: *spec,
	}

	operator.UpdateObject(deployment,
		operator.WithLabels(sb.Labels),
		operator.WithOwner(sb),
	)

	return deployment, nil
}

func makeStorageBrokerDeploymentSpec(sb *v1alpha1.StorageBroker, sbp *v1alpha1.StorageBrokerProfile) (*appsv1.DeploymentSpec, error) {
	cpf := sbp.Spec.CommonFields

	image := NeonDefaultImage
	if cpf.Image != nil {
		image = *cpf.Image
	}

	// Set replicas - use StorageBroker.Spec.Replicas if set, otherwise fall back to profile MinReplicas
	replicas := int32(1)
	if sb.Spec.Replicas != nil {
		replicas = *sb.Spec.Replicas
	} else if sbp.Spec.MinReplicas != nil {
		replicas = int32(*sbp.Spec.MinReplicas)
	}

	// Build pod labels
	labels := map[string]string{
		"app":       "storagebroker",
		"component": "storagebroker-deployment",
	}

	container := corev1.Container{
		Name:            "storagebroker",
		Image:           image,
		ImagePullPolicy: cpf.ImagePullPolicy,
		Resources:       cpf.Resources,
		Command:         []string{"storage_broker"},
		Args:            []string{"--listen-addr=0.0.0.0:50051", "--listen-https-addr=0.0.0.0:50052"},
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 50051,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "https",
				ContainerPort: 50052,
				Protocol:      corev1.ProtocolTCP,
			},
		},
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
		},
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
