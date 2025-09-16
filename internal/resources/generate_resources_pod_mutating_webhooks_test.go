// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"
	"testing"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateDeployment(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "test-namespace")
	os.Setenv("IMAGE_REGISTRY", "test-registry")
	os.Setenv("IMAGE_REPOSITORY", "test-repo")
	os.Setenv("APP_VERSION", "v1.0.0")

	class := overcommit.OvercommitClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-class",
		},
		Spec: overcommit.OvercommitClassSpec{
			Labels:      map[string]string{"key": "value"},
			Annotations: map[string]string{"annotation-key": "annotation-value"},
		},
	}

	deployment := CreateDeployment(class)

	if deployment.ObjectMeta.Name != "test-class-overcommit-webhook" {
		t.Errorf("Expected deployment name 'test-class-overcommit-webhook', got '%s'", deployment.ObjectMeta.Name)
	}

	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != 1 {
		t.Errorf("Expected replicas to be 1, got '%v'", deployment.Spec.Replicas)
	}
}

func TestCreateDeploymentWithTolerations(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "test-namespace")
	os.Setenv("IMAGE_REGISTRY", "test-registry")
	os.Setenv("IMAGE_REPOSITORY", "test-repo")
	os.Setenv("APP_VERSION", "v1.0.0")

	class := overcommit.OvercommitClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-class",
		},
		Spec: overcommit.OvercommitClassSpec{
			Labels:      map[string]string{"key": "value"},
			Annotations: map[string]string{"annotation-key": "annotation-value"},
			Tolerations: []corev1.Toleration{
				{
					Key:      "key1",
					Operator: corev1.TolerationOpExists,
				},
			},
		},
	}

	deployment := CreateDeployment(class)

	if deployment.ObjectMeta.Name != "test-class-overcommit-webhook" {
		t.Errorf("Expected deployment name 'test-class-overcommit-webhook', got '%s'", deployment.ObjectMeta.Name)
	}

	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != 1 {
		t.Errorf("Expected replicas to be 1, got '%v'", deployment.Spec.Replicas)
	}

	if len(deployment.Spec.Template.Spec.Tolerations) != 1 {
		t.Errorf("Expected 1 toleration, got '%d'", len(deployment.Spec.Template.Spec.Tolerations))
	}

	if deployment.Spec.Template.Spec.Tolerations[0].Key != "key1" {
		t.Errorf("Expected toleration key 'key1', got '%s'", deployment.Spec.Template.Spec.Tolerations[0].Key)
	}
}

func TestCreateService(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "test-namespace")

	service := CreateService("test-class")

	if service.ObjectMeta.Name != "test-class-webhook-service" {
		t.Errorf("Expected service name 'test-class-webhook-service', got '%s'", service.ObjectMeta.Name)
	}

	if service.Spec.Ports[0].Port != 443 {
		t.Errorf("Expected service port 443, got '%d'", service.Spec.Ports[0].Port)
	}
}

func TestCreateCertificate(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "test-namespace")

	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
	}

	certificate := CreateCertificate("test-class", service)

	if certificate.ObjectMeta.Name != "test-class-webhook-certificate" {
		t.Errorf("Expected certificate name 'test-class-webhook-certificate', got '%s'", certificate.ObjectMeta.Name)
	}

	if certificate.Spec.SecretName != "test-class-webhook-secret" {
		t.Errorf("Expected secret name 'test-class-webhook-secret', got '%s'", certificate.Spec.SecretName)
	}

	expectedDuration := 87600 * time.Hour
	if certificate.Spec.Duration.Duration != expectedDuration {
		t.Errorf("Expected duration '%v', got '%v'", expectedDuration, certificate.Spec.Duration.Duration)
	}
}

func TestResourceMustParse(t *testing.T) {
	memory := resourceMustParse("64Mi")
	if memory.String() != "64Mi" {
		t.Errorf("Expected '64Mi', got '%s'", memory.String())
	}

	cpu := resourceMustParse("250m")
	if cpu.String() != "250m" {
		t.Errorf("Expected '250m', got '%s'", cpu.String())
	}
}
