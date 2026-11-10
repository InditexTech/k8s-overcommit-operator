// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"os"
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

var _ = Describe("OvercommitClass Controller", func() {
	Context("When only tolerationSeconds changes in the OvercommitClass", func() {
		const (
			className = "toleration-seconds"
			timeout   = 10 * time.Second
			interval  = 250 * time.Millisecond
		)

		// The class reconciler generates one webhook Deployment per class and copies
		// the class tolerations into its pod template.
		deploymentName := className + "-overcommit-webhook"

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

		// deploymentTolerationSeconds reads tolerationSeconds from the pod template of
		// the class webhook Deployment. Negative sentinels distinguish the failure
		// modes: -1 the Deployment is missing, -2 it does not carry exactly one
		// toleration and -3 the toleration has no tolerationSeconds.
		deploymentTolerationSeconds := func() int64 {
			deployment := &appsv1.Deployment{}
			if err := k8sClient.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: os.Getenv("POD_NAMESPACE")}, deployment); err != nil {
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
			// A previous spec may have left the class behind.
			existing := &overcommitv1.OvercommitClass{ObjectMeta: metav1.ObjectMeta{Name: className}}
			err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, existing)
			if err == nil {
				Expect(k8sClient.Delete(ctx, existing)).To(Succeed())
			} else {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, &overcommitv1.OvercommitClass{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			err := k8sClient.Delete(ctx, &overcommitv1.OvercommitClass{ObjectMeta: metav1.ObjectMeta{Name: className}})
			if err != nil {
				Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
		})

		It("Should propagate the new value to the class webhook Deployment", func() {
			By("Creating an OvercommitClass with a toleration of 30 seconds")
			Expect(k8sClient.Create(ctx, &overcommitv1.OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{Name: className},
				Spec: overcommitv1.OvercommitClassSpec{
					CpuOvercommit:      0.5,
					MemoryOvercommit:   0.5,
					ExcludedNamespaces: "kube-system",
					Tolerations:        tolerationWithSeconds(30),
				},
			})).To(Succeed())

			By("Waiting for the webhook Deployment to carry tolerationSeconds 30")
			Eventually(func() int64 {
				return deploymentTolerationSeconds()
			}, timeout, interval).Should(Equal(int64(30)), "the class webhook Deployment did not get the initial toleration")

			By("Changing only tolerationSeconds in the OvercommitClass")
			Eventually(func() error {
				overcommitClass := &overcommitv1.OvercommitClass{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, overcommitClass); err != nil {
					return err
				}
				overcommitClass.Spec.Tolerations = tolerationWithSeconds(600)
				return k8sClient.Update(ctx, overcommitClass)
			}, timeout, interval).Should(Succeed())

			By("Expecting the webhook Deployment to get the new tolerationSeconds")
			Eventually(func() int64 {
				return deploymentTolerationSeconds()
			}, timeout, interval).Should(Equal(int64(600)), "the class webhook Deployment kept the old tolerationSeconds")
		})
	})
})
