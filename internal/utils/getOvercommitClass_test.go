// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GetOvercommitClassSpec", func() {
	var overcommitClass *overcommit.OvercommitClass

	BeforeEach(func() {
		// Create a test OvercommitClass with IsDefault: false
		overcommitClass = &overcommit.OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-overcommitclass",
			},
			Spec: overcommit.OvercommitClassSpec{
				CpuOvercommit:      0.5,
				MemoryOvercommit:   0.5,
				ExcludedNamespaces: "kube-system",
				IsDefault:          false,
			},
		}

		// Create the resource in the test cluster
		err := k8sClient.Create(context.Background(), overcommitClass)
		Expect(err).NotTo(HaveOccurred(), "Failed to create OvercommitClass")
	})

	AfterEach(func() {
		// Delete the resource after each test
		err := k8sClient.Delete(context.Background(), overcommitClass)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete OvercommitClass")
	})

	It("should retrieve the OvercommitClassSpec correctly", func() {
		// Try the GetOvercommitClassSpec function
		spec, err := GetOvercommitClassSpec(context.TODO(), "test-overcommitclass", k8sClient)
		Expect(err).NotTo(HaveOccurred(), "Failed to get OvercommitClassSpec")
		Expect(spec).NotTo(BeNil(), "Spec should not be nil")
		Expect(spec).To(Equal(&overcommitClass.Spec), "Spec should match the created OvercommitClass")
	})
})

var _ = Describe("GetDefaultSpec", func() {
	var overcommitClass *overcommit.OvercommitClass

	BeforeEach(func() {
		// Create a test OvercommitClass with IsDefault: true
		overcommitClass = &overcommit.OvercommitClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-default-overcommitclass",
			},
			Spec: overcommit.OvercommitClassSpec{
				CpuOvercommit:      0.5,
				MemoryOvercommit:   0.5,
				ExcludedNamespaces: "kube-system",
				IsDefault:          true,
			},
		}

		// Create the resource in the test cluster
		err := k8sClient.Create(context.Background(), overcommitClass)
		Expect(err).NotTo(HaveOccurred(), "Failed to create default OvercommitClass")
	})

	AfterEach(func() {
		// Delete the resource after each test
		err := k8sClient.Delete(context.Background(), overcommitClass)
		Expect(err).NotTo(HaveOccurred(), "Failed to delete default OvercommitClass")
	})

	It("should retrieve the default OvercommitClassSpec correctly", func() {
		// Try the GetDefaultSpec function
		spec, err := GetDefaultSpec(context.Background(), k8sClient)
		Expect(err).NotTo(HaveOccurred(), "Failed to get default OvercommitClassSpec")
		Expect(spec).NotTo(BeNil(), "Spec should not be nil")
		Expect(spec.IsDefault).To(BeTrue(), "Spec.IsDefault should be true")
	})
})
