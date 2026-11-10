// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"time"

	overcommitv1 "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Overcommit Controller", func() {
	Context("When only tolerationSeconds changes in the Overcommit CR", func() {
		const (
			overcommitName      = "cluster"
			overcommitNamespace = "k8s-overcommit"
			timeout             = 10 * time.Second
			interval            = 250 * time.Millisecond
		)

		// Deployments owned by the reconciler whose pod template gets the
		// tolerations of the Overcommit CR.
		managedDeployments := []string{
			"k8s-overcommit-class-validating-webhook",
			"k8s-overcommit-pod-validating-webhook",
			"k8s-overcommit-overcommitclass-controller",
		}

		// tolerationWithSeconds builds the not-ready/unreachable style toleration
		// the operator is expected to copy verbatim into the managed Deployments.
		tolerationWithSeconds := func(seconds int64) []corev1.Toleration {
			s := seconds
			return []corev1.Toleration{
				{
					Key:               "node.kubernetes.io/not-ready",
					Operator:          corev1.TolerationOpExists,
					Effect:            corev1.TaintEffectNoExecute,
					TolerationSeconds: &s,
				},
			}
		}

		// deploymentTolerationSeconds reads tolerationSeconds from the pod template
		// of a managed Deployment. Negative sentinels distinguish the failure modes:
		// -1 the Deployment is missing, -2 it does not carry exactly one toleration
		// and -3 the toleration has no tolerationSeconds.
		deploymentTolerationSeconds := func(name string) int64 {
			deployment := &appsv1.Deployment{}
			if err := k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: overcommitNamespace}, deployment); err != nil {
				return -1
			}
			tolerations := deployment.Spec.Template.Spec.Tolerations
			if len(tolerations) != 1 {
				return -2
			}
			if tolerations[0].TolerationSeconds == nil {
				return -3
			}
			return *tolerations[0].TolerationSeconds
		}

		BeforeEach(func() {
			// The namespace is also created by the "create related resources" spec;
			// tolerate it being there already.
			namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: overcommitNamespace}}
			if err := k8sClient.Create(ctx, namespace); err != nil && !apierrors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}

			// Overcommit is a cluster-scoped singleton, so the other spec may have
			// left it behind. Delete it and wait until it is gone before creating it.
			existing := &overcommitv1.Overcommit{}
			err := k8sClient.Get(ctx, client.ObjectKey{Name: overcommitName}, existing)
			if err == nil {
				Expect(k8sClient.Delete(ctx, existing)).To(Succeed())
			} else {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: overcommitName}, &overcommitv1.Overcommit{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			overcommit := &overcommitv1.Overcommit{ObjectMeta: metav1.ObjectMeta{Name: overcommitName}}
			err := k8sClient.Delete(ctx, overcommit)
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
		})

		It("Should propagate the new value to the managed Deployments", func() {
			By("Creating an Overcommit CR with a toleration of 30 seconds")
			Expect(k8sClient.Create(ctx, &overcommitv1.Overcommit{
				ObjectMeta: metav1.ObjectMeta{Name: overcommitName},
				Spec: overcommitv1.OvercommitSpec{
					OvercommitLabel: "test-overcommit",
					Tolerations:     tolerationWithSeconds(30),
				},
			})).To(Succeed())

			By("Waiting for the managed Deployments to carry tolerationSeconds 30")
			for _, name := range managedDeployments {
				Eventually(func() int64 {
					return deploymentTolerationSeconds(name)
				}, timeout, interval).Should(Equal(int64(30)), "deployment %s did not get the initial toleration", name)
			}

			By("Changing only tolerationSeconds in the Overcommit CR")
			Eventually(func() error {
				overcommit := &overcommitv1.Overcommit{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: overcommitName}, overcommit); err != nil {
					return err
				}
				overcommit.Spec.Tolerations = tolerationWithSeconds(600)
				return k8sClient.Update(ctx, overcommit)
			}, timeout, interval).Should(Succeed())

			By("Expecting the managed Deployments to get the new tolerationSeconds")
			for _, name := range managedDeployments {
				Eventually(func() int64 {
					return deploymentTolerationSeconds(name)
				}, timeout, interval).Should(Equal(int64(600)), "deployment %s kept the old tolerationSeconds", name)
			}
		})
	})
})
