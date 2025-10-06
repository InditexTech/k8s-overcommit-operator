// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateOvercommitClassControllerDeployment(overcommitObject overcommit.Overcommit) *appsv1.Deployment {
	replicas := int32(1)
	labels := overcommitObject.Spec.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "overcommit-controller"

	// Ensure annotations are properly handled
	annotations := overcommitObject.Spec.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-overcommit-overcommitclass-controller",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "overcommit-controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: os.Getenv("SERVICE_ACCOUNT_NAME"),
					Containers: []corev1.Container{
						{
							Name:  "overcommit-controller",
							Image: os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION"),
							Args: []string{
								"--metrics-bind-address=:8080",
								"-metrics-secure=false",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENABLE_OVERCOMMIT_CLASS_CONTROLLER",
									Value: "true",
								},
								{
									Name:  "IMAGE_REGISTRY",
									Value: os.Getenv("IMAGE_REGISTRY"),
								},
								{
									Name:  "IMAGE_REPOSITORY",
									Value: os.Getenv("IMAGE_REPOSITORY"),
								},
								{
									Name:  "APP_VERSION",
									Value: os.Getenv("APP_VERSION"),
								},
								{
									Name:  "POD_NAMESPACE",
									Value: os.Getenv("POD_NAMESPACE"),
								},
								{
									Name:  "SERVICE_ACCOUNT_NAME",
									Value: os.Getenv("SERVICE_ACCOUNT_NAME"),
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080, Name: "metrics", Protocol: corev1.ProtocolTCP},
							},
						},
					},
					NodeSelector: overcommitObject.Spec.NodeSelector,
					Tolerations:  overcommitObject.Spec.Tolerations,
				},
			},
		},
	}
}
