// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"time"

	overcommitv1 "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OvercommitClass failurePolicy", func() {
	const (
		className = "failure-policy"
		timeout   = 10 * time.Second
		interval  = 250 * time.Millisecond
	)

	// The class reconciler generates one MutatingWebhookConfiguration per class.
	webhookConfigName := className + "-overcommit-webhook"

	// webhookFailurePolicies returns the failure policy of every webhook in the
	// class MutatingWebhookConfiguration, or nil when the object does not exist.
	webhookFailurePolicies := func() []admissionv1.FailurePolicyType {
		webhookConfig := &admissionv1.MutatingWebhookConfiguration{}
		if err := k8sClient.Get(ctx, client.ObjectKey{Name: webhookConfigName}, webhookConfig); err != nil {
			return nil
		}
		policies := make([]admissionv1.FailurePolicyType, 0, len(webhookConfig.Webhooks))
		for _, webhook := range webhookConfig.Webhooks {
			if webhook.FailurePolicy == nil {
				policies = append(policies, "")
				continue
			}
			policies = append(policies, *webhook.FailurePolicy)
		}
		return policies
	}

	// classSpec returns a minimal valid spec, leaving failurePolicy untouched.
	classSpec := func() overcommitv1.OvercommitClassSpec {
		return overcommitv1.OvercommitClassSpec{
			CpuOvercommit:      0.5,
			MemoryOvercommit:   0.5,
			ExcludedNamespaces: "kube-system",
		}
	}

	// The class is cluster-scoped and envtest runs no garbage collector, so the
	// generated MutatingWebhookConfiguration is removed by hand.
	cleanUp := func() {
		overcommitClass := &overcommitv1.OvercommitClass{ObjectMeta: metav1.ObjectMeta{Name: className}}
		err := k8sClient.Delete(ctx, overcommitClass)
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}

		webhookConfig := &admissionv1.MutatingWebhookConfiguration{ObjectMeta: metav1.ObjectMeta{Name: webhookConfigName}}
		err = k8sClient.Delete(ctx, webhookConfig)
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}
	}

	BeforeEach(func() {
		cleanUp()
		Eventually(func() bool {
			err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, &overcommitv1.OvercommitClass{})
			return apierrors.IsNotFound(err)
		}, timeout, interval).Should(BeTrue())
	})

	AfterEach(func() {
		cleanUp()
	})

	It("Should default to Fail when the field is not set", func() {
		By("Creating an OvercommitClass without failurePolicy")
		overcommitClass := &overcommitv1.OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{Name: className},
			Spec:       classSpec(),
		}
		Expect(k8sClient.Create(ctx, overcommitClass)).To(Succeed())

		By("Expecting the API server to apply the Fail default")
		Expect(overcommitClass.Spec.FailurePolicy).To(Equal(overcommitv1.FailurePolicyFail))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: className}, overcommitClass)).To(Succeed())
		Expect(overcommitClass.Spec.FailurePolicy).To(Equal(overcommitv1.FailurePolicyFail))

		By("Expecting the generated MutatingWebhookConfiguration to use Fail")
		Eventually(webhookFailurePolicies, timeout, interval).Should(ConsistOf(admissionv1.Fail))
	})

	It("Should use Ignore when the class asks for it", func() {
		By("Creating an OvercommitClass with failurePolicy Ignore")
		spec := classSpec()
		spec.FailurePolicy = overcommitv1.FailurePolicyIgnore
		Expect(k8sClient.Create(ctx, &overcommitv1.OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{Name: className},
			Spec:       spec,
		})).To(Succeed())

		By("Expecting the generated MutatingWebhookConfiguration to use Ignore")
		Eventually(webhookFailurePolicies, timeout, interval).Should(ConsistOf(admissionv1.Ignore))
	})

	It("Should propagate a change on an existing class", func() {
		By("Creating an OvercommitClass that relies on the Fail default")
		Expect(k8sClient.Create(ctx, &overcommitv1.OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{Name: className},
			Spec:       classSpec(),
		})).To(Succeed())
		Eventually(webhookFailurePolicies, timeout, interval).Should(ConsistOf(admissionv1.Fail))

		By("Switching the class to Ignore")
		Eventually(func() error {
			overcommitClass := &overcommitv1.OvercommitClass{}
			if err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, overcommitClass); err != nil {
				return err
			}
			overcommitClass.Spec.FailurePolicy = overcommitv1.FailurePolicyIgnore
			return k8sClient.Update(ctx, overcommitClass)
		}, timeout, interval).Should(Succeed())

		By("Expecting the existing MutatingWebhookConfiguration to be updated")
		Eventually(webhookFailurePolicies, timeout, interval).Should(ConsistOf(admissionv1.Ignore))

		By("Switching the class back to Fail")
		Eventually(func() error {
			overcommitClass := &overcommitv1.OvercommitClass{}
			if err := k8sClient.Get(ctx, client.ObjectKey{Name: className}, overcommitClass); err != nil {
				return err
			}
			overcommitClass.Spec.FailurePolicy = overcommitv1.FailurePolicyFail
			return k8sClient.Update(ctx, overcommitClass)
		}, timeout, interval).Should(Succeed())

		Eventually(webhookFailurePolicies, timeout, interval).Should(ConsistOf(admissionv1.Fail))
	})

	It("Should reject a value that is not Fail or Ignore", func() {
		for _, invalid := range []string{"fail", "failure", "Unknown"} {
			By("Creating an OvercommitClass with failurePolicy " + invalid)
			spec := classSpec()
			spec.FailurePolicy = overcommitv1.FailurePolicyType(invalid)
			err := k8sClient.Create(ctx, &overcommitv1.OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{Name: className},
				Spec:       spec,
			})

			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "expected an invalid error, got: %v", err)
			Expect(err.Error()).To(ContainSubstring("failurePolicy"))

			By("Expecting no MutatingWebhookConfiguration to be generated")
			Consistently(webhookFailurePolicies, 1*time.Second, interval).Should(BeEmpty())
		}
	})
})
