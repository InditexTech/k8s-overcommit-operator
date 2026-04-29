// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodCustomDefaulter Webhook", func() {
	var defaulter *PodCustomDefaulter

	BeforeEach(func() {
		defaulter = &PodCustomDefaulter{}
		defaulter.InjectClient(k8sClient)
		defaulter.InjectRecorder(recorder)
	})

	Context("Defaulting webhook", func() {
		It("Should apply overcommit defaults to a Pod", func() {
			// Create a sample Pod object
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"inditex.com/overcommit-class": "default-overcommitclass",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			}

			// Call the Default method
			err := defaulter.Default(context.TODO(), pod)
			Expect(err).NotTo(HaveOccurred())

			// Verify that the Overcommit function was applied
			expectedRequests := corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			}

			actualRequests := pod.Spec.Containers[0].Resources.Requests

			Expect(actualRequests[corev1.ResourceCPU].Equal(expectedRequests[corev1.ResourceCPU])).To(BeTrue())
			Expect(actualRequests[corev1.ResourceMemory].Equal(expectedRequests[corev1.ResourceMemory])).To(BeTrue())
		})

		It("Should handle non-Pod gracefully", func() {
			// With typed generics, passing a non-Pod is now a compile-time error.
			// This test is no longer applicable.
		})
	})
})
