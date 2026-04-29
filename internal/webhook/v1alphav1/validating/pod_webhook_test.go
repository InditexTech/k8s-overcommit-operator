// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PodCustomValidator Webhook", func() {
	var (
		validator *PodCustomValidator
		ctx       context.Context
		pod       *corev1.Pod
	)

	BeforeEach(func() {
		validator = &PodCustomValidator{}
		validator.Client = k8sClient
		ctx = context.Background()

		// Create a sample Pod object
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test-pod",
				Labels: map[string]string{"inditex.com/overcommit-class": "default-overcommitclass"},
			},
		}
	})

	AfterEach(func() {
		os.Unsetenv("LABEL_OVERCOMMIT_CLASS")
	})

	Context("ValidateCreate", func() {
		It("should pass validation when Pod has valid overcommit class label", func() {
			// Validate Pod creation
			warnings, err := validator.ValidateCreate(ctx, pod)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail when Pod lacks overcommit class label", func() {
			delete(pod.Labels, "inditex.com/overcommit-class")

			// Validate Pod creation
			warnings, err := validator.ValidateCreate(ctx, pod)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Pod without overcommit class"))
		})

		It("should fail validation when OvercommitClass doesnt exists", func() {
			// Validate Pod creation
			pod.Labels["inditex.com/overcommit-class"] = "nonexistent-overcommitclass"
			warnings, err := validator.ValidateCreate(ctx, pod)
			Expect(warnings).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ValidateUpdate", func() {
		It("should pass validation for Pod update", func() {
			// Validate Pod update
			warnings, err := validator.ValidateUpdate(ctx, pod, pod)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ValidateDelete", func() {
		It("should pass validation for Pod deletion", func() {
			// Validate Pod deletion
			warnings, err := validator.ValidateDelete(ctx, pod)
			Expect(warnings).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
