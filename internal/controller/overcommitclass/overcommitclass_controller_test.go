// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	certmanager "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("OvercommitClass Controller", func() {
	Context("When creating an OvercommitClass resource", func() {
		AfterEach(func() {
			// Clean up resources created during the test
			Expect(k8sClient.Delete(ctx, &overcommit.OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			})).To(Succeed())

			Expect(k8sClient.Delete(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: os.Getenv("POD_NAMESPACE"),
				},
			})).To(Succeed())
		})

		It("Should create dependent resources", func() {
			// Define a sample OvercommitClass resource
			overcommitClass := &overcommit.OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: overcommit.OvercommitClassSpec{
					IsDefault:          false,
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
				},
			}

			// Create the namespace
			Expect(k8sClient.Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: os.Getenv("POD_NAMESPACE"),
				},
			}))

			// Create the OvercommitClass resource
			Expect(k8sClient.Create(ctx, overcommitClass)).To(Succeed())

			// Verify that dependent resources are created
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommit-webhook", Namespace: os.Getenv("POD_NAMESPACE")}, deployment)
			}, 10*time.Second, 250*time.Millisecond).Should(Succeed())

			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-webhook-service", Namespace: os.Getenv("POD_NAMESPACE")}, service)
			}, 10*time.Second, 250*time.Millisecond).Should(Succeed())

			certificate := &certmanager.Certificate{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-webhook-certificate", Namespace: os.Getenv("POD_NAMESPACE")}, certificate)
			}, 10*time.Second, 250*time.Millisecond).Should(Succeed())

			webhookConfig := &admissionv1.MutatingWebhookConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommit-webhook"}, webhookConfig)
			}, 10*time.Second, 250*time.Millisecond).Should(Succeed())
		})
	})
})

var _ = Describe("OvercommitClass Controller", func() {
	Context("When deleting an OvercommitClass resource", func() {
		It("Should clean up dependent resources", func() {
			// Define a sample OvercommitClass resource
			overcommitClass := &overcommit.OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-overcommitclass-2",
				},
				Spec: overcommit.OvercommitClassSpec{
					IsDefault:          true,
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
				},
			}

			// Create the OvercommitClass resource
			Expect(k8sClient.Create(ctx, overcommitClass)).To(Succeed())

			// Delete the OvercommitClass resource
			Expect(k8sClient.Delete(ctx, overcommitClass)).To(Succeed())

			// Verify that dependent resources are deleted
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommitclass-2-webhook-deployment", Namespace: os.Getenv("POD_NAMESPACE")}, deployment)
			}).ShouldNot(Succeed())

			service := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommitclass-2-webhook-service", Namespace: os.Getenv("POD_NAMESPACE")}, service)
			}).ShouldNot(Succeed())

			certificate := &certmanager.Certificate{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommitclass-2-webhook-certificate", Namespace: os.Getenv("POD_NAMESPACE")}, certificate)
			}).ShouldNot(Succeed())

			webhookConfig := &admissionv1.MutatingWebhookConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: "test-overcommitclass-2-overcommit-webhook"}, webhookConfig)
			}).ShouldNot(Succeed())
		})
	})
})
