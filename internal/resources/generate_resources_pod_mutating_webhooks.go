// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateDeployment(class overcommit.OvercommitClass) *appsv1.Deployment {
	replicas := int32(1)

	if class.Spec.Labels == nil {
		class.Spec.Labels = make(map[string]string)
	}

	labels := class.Spec.Labels
	labels["app"] = class.ObjectMeta.Name + "-overcommit-webhook"

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      class.ObjectMeta.Name + "-overcommit-webhook",
			Namespace: os.Getenv("POD_NAMESPACE"),
			Labels: map[string]string{
				"app": class.ObjectMeta.Name + "-overcommit-webhook",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": class.ObjectMeta.Name + "-overcommit-webhook",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: class.Spec.Annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: os.Getenv("SERVICE_ACCOUNT_NAME"),
					Containers: []corev1.Container{
						{
							Name:    "k8s-overcommit",
							Image:   os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION"),
							Command: []string{"/manager"},
							Args: []string{
								"--metrics-bind-address=:8080",
								"-metrics-secure=false",
							},
							Env: []corev1.EnvVar{
								{Name: "APP_VERSION", Value: os.Getenv("APP_VERSION")},
								{Name: "WEBHOOK_CERT_DIR", Value: "/etc/webhook/config"},
								{Name: "ENABLE_CONTROLLER", Value: "false"},
								{Name: "ENABLE_POD_MUTATING_WEBHOOK", Value: "true"},
								{Name: "OVERCOMMIT_CLASS_NAME", Value: class.Name},
								{Name: "SERVICE_ACCOUNT_NAME", Value: os.Getenv("SERVICE_ACCOUNT_NAME")},
								{Name: "POD_NAMESPACE", Value: os.Getenv("POD_NAMESPACE")},
								{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.name"},
								}},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 9443},
								{ContainerPort: 8080, Name: "metrics", Protocol: corev1.ProtocolTCP},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resourceMustParse("64Mi"),
									corev1.ResourceCPU:    resourceMustParse("250m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resourceMustParse("4Gi"),
									corev1.ResourceCPU:    resourceMustParse("2"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "webhook-tls-secret",
									MountPath: "/etc/webhook/config",
									ReadOnly:  true,
								},
							},
						},
					},
					NodeSelector: class.Spec.NodeSelector,
					Tolerations:  class.Spec.Tolerations,
					Volumes: []corev1.Volume{
						{
							Name: "webhook-tls-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: class.Name + "-webhook-secret",
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceMustParse(value string) resource.Quantity {
	res := resource.MustParse(value)
	return res
}

func CreateService(name string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-webhook-service",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name + "-overcommit-webhook"},
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					TargetPort: intstr.FromInt(9443),
				},
			},
		},
	}
}

func CreateCertificate(name string, svc corev1.Service) *certmanager.Certificate {
	return &certmanager.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-webhook-certificate",
			Namespace: os.Getenv("POD_NAMESPACE"),
		},
		Spec: certmanager.CertificateSpec{
			SecretName: name + "-webhook-secret",
			Duration: &metav1.Duration{
				Duration: 87600 * time.Hour,
			},
			RenewBefore: &metav1.Duration{
				Duration: 720 * time.Hour,
			},
			DNSNames: []string{
				svc.Name + "." + svc.Namespace + ".svc",
				svc.Name + "." + svc.Namespace + ".svc.cluster.local",
			},
			IssuerRef: certmanagermeta.ObjectReference{
				Name: "k8s-overcommit-issuer",
				Kind: "Issuer",
			},
		},
	}
}

func getSelectorClassNotExist(label string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      label,
				Operator: metav1.LabelSelectorOpDoesNotExist,
			},
		},
	}
}

func getMatchCondition(isDefault bool, name string, excludedNamespaces string, label string) []admissionv1.MatchCondition {
	matchConditions := []admissionv1.MatchCondition{}
	matchConditions = append(matchConditions, admissionv1.MatchCondition{
		Name:       "exclude-namespaces",
		Expression: "!object.metadata.namespace.matches('" + excludedNamespaces + "')",
	})

	return matchConditions
}

func getObjectSelector(isDefault bool, label string, name string) *metav1.LabelSelector {
	if isDefault {
		return getSelectorClassNotExist(label)
	} else {
		return &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      label,
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{name},
				},
			},
		}
	}
}

func CreateMutatingWebhookConfiguration(class overcommit.OvercommitClass, svc corev1.Service, cert certmanager.Certificate, label string) *admissionv1.MutatingWebhookConfiguration {

	var path = "/mutate--v1-pod"
	var scope = admissionv1.NamespacedScope
	var policy = admissionv1.Fail
	var sideEffect = admissionv1.SideEffectClassNone

	webhookConfig := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: class.Name + "-overcommit-webhook",
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from": cert.ObjectMeta.Namespace + "/" + cert.Name,
			},
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name: class.ObjectMeta.Name + "-overcommit.inditex.dev",
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      svc.Name,
						Namespace: svc.Namespace,
						Path:      &path,
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
						},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
							Scope:       &scope,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				FailurePolicy:           &policy,
				SideEffects:             &sideEffect,
				MatchConditions:         getMatchCondition(false, class.Name, class.Spec.ExcludedNamespaces, label),
				ObjectSelector:          getObjectSelector(false, label, class.Name),
			},
		},
	}

	if class.Spec.IsDefault {
		webhookConfig.Webhooks = append(webhookConfig.Webhooks, admissionv1.MutatingWebhook{
			Name: "default-" + class.ObjectMeta.Name + "-overcommit.inditex.dev",
			ClientConfig: admissionv1.WebhookClientConfig{
				Service: &admissionv1.ServiceReference{
					Name:      svc.Name,
					Namespace: svc.Namespace,
					Path:      &path,
				},
			},
			Rules: []admissionv1.RuleWithOperations{
				{
					Operations: []admissionv1.OperationType{
						admissionv1.Create,
					},
					Rule: admissionv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
						Scope:       &scope,
					},
				},
			},
			AdmissionReviewVersions: []string{"v1"},
			FailurePolicy:           &policy,
			SideEffects:             &sideEffect,
			MatchConditions:         getMatchCondition(class.Spec.IsDefault, class.Name, class.Spec.ExcludedNamespaces, label),
			ObjectSelector:          getObjectSelector(class.Spec.IsDefault, label, class.Name),
		})
	}
	return webhookConfig
}
