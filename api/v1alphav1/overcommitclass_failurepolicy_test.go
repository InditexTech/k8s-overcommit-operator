// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These specs exercise the CRD schema itself, not the validating webhook: the
// enum and the default are enforced by the API server at admission time.
var _ = Describe("OvercommitClass failurePolicy schema", func() {
	const className = "failure-policy-schema"

	validSpec := func() OvercommitClassSpec {
		return OvercommitClassSpec{
			CpuOvercommit:      0.5,
			MemoryOvercommit:   0.5,
			ExcludedNamespaces: "kube-system",
		}
	}

	BeforeEach(func() {
		existing := &OvercommitClass{ObjectMeta: metav1.ObjectMeta{Name: className}}
		err := k8sClient.Delete(context.TODO(), existing)
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}
	})

	AfterEach(func() {
		err := k8sClient.Delete(context.TODO(), &OvercommitClass{ObjectMeta: metav1.ObjectMeta{Name: className}})
		if err != nil {
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}
	})

	It("Should default to Fail when the field is omitted", func() {
		overcommitClass := &OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{Name: className},
			Spec:       validSpec(),
		}
		Expect(k8sClient.Create(context.TODO(), overcommitClass)).To(Succeed())

		Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: className}, overcommitClass)).To(Succeed())
		Expect(overcommitClass.Spec.FailurePolicy).To(Equal(FailurePolicyFail))
	})

	DescribeTable("Should accept the allowed values",
		func(policy FailurePolicyType) {
			overcommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{Name: className},
				Spec:       validSpec(),
			}
			overcommitClass.Spec.FailurePolicy = policy
			Expect(k8sClient.Create(context.TODO(), overcommitClass)).To(Succeed())

			stored := &OvercommitClass{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: className}, stored)).To(Succeed())
			Expect(stored.Spec.FailurePolicy).To(Equal(policy))
		},
		Entry("Fail", FailurePolicyFail),
		Entry("Ignore", FailurePolicyIgnore),
	)

	DescribeTable("Should reject any other value",
		func(policy string) {
			overcommitClass := &OvercommitClass{
				ObjectMeta: metav1.ObjectMeta{Name: className},
				Spec:       validSpec(),
			}
			overcommitClass.Spec.FailurePolicy = FailurePolicyType(policy)

			err := k8sClient.Create(context.TODO(), overcommitClass)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "expected an invalid error, got: %v", err)
			Expect(err.Error()).To(ContainSubstring("failurePolicy"))
			Expect(err.Error()).To(ContainSubstring("Unsupported value"))
		},
		Entry("lowercase fail", "fail"),
		Entry("failure", "failure"),
		Entry("unknown", "Unknown"),
	)
})
