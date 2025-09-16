// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GenerateCertificateValidatingPods(issuer certmanagerv1.Issuer, svc corev1.Service) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-validating-webhook",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: "pod-validating-webhook",
			IssuerRef: certmanagermeta.ObjectReference{
				Name: issuer.Name,
			},
			Duration: &metav1.Duration{
				Duration: 365 * 24 * time.Hour,
			},
			RenewBefore: &metav1.Duration{
				Duration: 30 * 24 * time.Hour,
			},
			DNSNames: []string{
				svc.Name + "." + svc.Namespace + ".svc",
				svc.Name + "." + svc.Namespace + ".svc.cluster.local",
			},
		},
	}
}

func GeneratePodValidatingDeployment(overcommitObject overcommit.Overcommit) *appsv1.Deployment {
	replicas := int32(1)
	labels := overcommitObject.Spec.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "k8s-overcommit-pod-validating-webhook"
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-overcommit-pod-validating-webhook",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "k8s-overcommit-pod-validating-webhook",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: overcommitObject.Spec.Annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: os.Getenv("SERVICE_ACCOUNT_NAME"),
					Containers: []corev1.Container{
						{
							Name:  "k8s-overcommit-pod-validating-webhook",
							Image: os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION"),
							Args: []string{
								"--metrics-bind-address=:8080",
								"-metrics-secure=false",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENABLE_POD_VALIDATING_WEBHOOK",
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
									Name:  "WEBHOOK_CERT_DIR",
									Value: "/etc/webhook/config",
								},
								{
									Name:  "SERVICE_ACCOUNT_NAME",
									Value: os.Getenv("SERVICE_ACCOUNT_NAME"),
								},
								{
									Name:  "POD_NAMESPACE",
									Value: os.Getenv("POD_NAMESPACE"),
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name"},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9443},
								{ContainerPort: 8080, Name: "metrics", Protocol: corev1.ProtocolTCP},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "webhook-cert",
									MountPath: "/etc/webhook/config",
									ReadOnly:  true,
								},
							},
						},
					},
					Tolerations: overcommitObject.Spec.Tolerations,
					Volumes: []corev1.Volume{
						{
							Name: "webhook-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "pod-validating-webhook",
								},
							},
						},
					},
				},
			},
		},
	}
}

func GeneratePodValidatingService(deployment appsv1.Deployment) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name + "-service",
			Namespace: deployment.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deployment.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(9443),
				},
			},
		},
	}
}

func GeneratePodValidatingWebhookConfiguration(deployment appsv1.Deployment, service corev1.Service, certificate certmanagerv1.Certificate, label string) *admissionv1.ValidatingWebhookConfiguration {
	var policy = admissionv1.Fail
	var sideEffects = admissionv1.SideEffectClassNone
	var path = "/validate--v1-pod"
	return &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.Name,
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": certificate.Namespace + "/" + certificate.Name,
			},
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name: "podvalidation.overcommit.inditex.dev",
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      service.Name,
						Namespace: service.Namespace,
						Path:      &path,
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{"CREATE", "UPDATE"},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
				FailurePolicy:           &policy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1"},
				ObjectSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      label,
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				MatchConditions: []admissionv1.MatchCondition{
					{
						Name:       "exclude-operator-namespace",
						Expression: "!object.metadata.namespace.matches('" + os.Getenv("POD_NAMESPACE") + "')",
					},
				},
			},
		},
	}
}

func GenerateOvercommitClassValidatingDeployment(overcommitObject overcommit.Overcommit) *appsv1.Deployment {
	replicas := int32(1)
	labels := overcommitObject.Spec.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "k8s-overcommit-class-validating-webhook"
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-overcommit-class-validating-webhook",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "k8s-overcommit-class-validating-webhook",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: overcommitObject.Spec.Annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: os.Getenv("SERVICE_ACCOUNT_NAME"),
					Containers: []corev1.Container{
						{
							Name:  "k8s-overcommit-class-validating-webhook",
							Image: os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION"),
							Args: []string{
								"--metrics-bind-address=:8080",
								"-metrics-secure=false",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENABLE_OC_VALIDATING_WEBHOOK",
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
									Name:  "WEBHOOK_CERT_DIR",
									Value: "/etc/webhook/config",
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name"},
									},
								},
								{
									Name:  "POD_NAMESPACE",
									Value: os.Getenv("POD_NAMESPACE"),
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9443},
								{ContainerPort: 8080, Name: "metrics", Protocol: corev1.ProtocolTCP},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "webhook-cert",
									MountPath: "/etc/webhook/config",
									ReadOnly:  true,
								},
							},
						},
					},
					Tolerations: overcommitObject.Spec.Tolerations,
					Volumes: []corev1.Volume{
						{
							Name: "webhook-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "oc-validating-webhook",
								},
							},
						},
					},
				},
			},
		},
	}
}

func GenerateOvercommitClassValidatingService(deployment appsv1.Deployment) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name + "-service",
			Namespace: deployment.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deployment.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(9443),
				},
			},
		},
	}
}

func GenerateCertificateValidatingOvercommitClass(issuer certmanagerv1.Issuer, svc corev1.Service) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oc-validating-webhook",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: "oc-validating-webhook",
			IssuerRef: certmanagermeta.ObjectReference{
				Name: issuer.Name,
			},
			Duration: &metav1.Duration{
				Duration: 365 * 24 * time.Hour,
			},
			RenewBefore: &metav1.Duration{
				Duration: 30 * 24 * time.Hour,
			},
			DNSNames: []string{
				svc.Name + "." + svc.Namespace + ".svc",
				svc.Name + "." + svc.Namespace + ".svc.cluster.local",
			},
		},
	}
}

func GenerateOvercommitClassValidatingWebhookConfiguration(deployment appsv1.Deployment, service corev1.Service, certificate certmanagerv1.Certificate) *admissionv1.ValidatingWebhookConfiguration {
	var policy = admissionv1.Fail
	var sideEffects = admissionv1.SideEffectClassNone
	var path = "/validate-overcommit-inditex-dev-v1alphav1-overcommitclass"

	return &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.Name,
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": certificate.Namespace + "/" + certificate.Name,
			},
		},
		Webhooks: []admissionv1.ValidatingWebhook{
			{
				Name: "overcommitclass.overcommit.inditex.dev",
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      service.Name,
						Namespace: service.Namespace,
						Path:      &path,
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{"CREATE", "UPDATE", "DELETE"},
						Rule: admissionv1.Rule{
							APIGroups:   []string{"overcommit.inditex.dev"},
							APIVersions: []string{"v1alphav1"},
							Resources:   []string{"overcommitclasses"},
						},
					},
				},
				FailurePolicy:           &policy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1"},
			},
		},
	}
}
