// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Overcommit", func() {
	var (
		pod              *corev1.Pod
		expectedRequests = corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
			corev1.ResourceMemory: *resource.NewQuantity(536870912, resource.BinarySI),
		}
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Labels: map[string]string{
					"inditex.com/overcommit-class": "test-class",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "test-container",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{},
						},
					},
				},
			},
		}
	})

	Describe("mutateContainers", func() {
		It("should mutate container requests based on overcommit values", func() {
			mutateContainers(pod.Spec.Containers, pod, 0.5, 0.5)

			Expect(pod.Spec.Containers[0].Resources.Requests).To(Equal(expectedRequests))
		})

		It("should not mutate containers if limits are nil", func() {
			pod.Spec.Containers[0].Resources.Limits = nil
			mutateContainers(pod.Spec.Containers, pod, 0.5, 0.5)

			Expect(pod.Spec.Containers[0].Resources.Requests).To(BeEmpty())
		})
	})

	Describe("makeOvercommit", func() {
		It("should apply overcommit to containers", func() {
			Overcommit(pod, recorder, k8sClient)

			Expect(pod.Spec.Containers[0].Resources.Requests).To(Equal(expectedRequests))
		})
	})

	Describe("Overcommit", func() {
		BeforeEach(func() {
			os.Setenv("OVERCOMMIT_CLASS_NAME", "test-class")
		})

		AfterEach(func() {
			os.Unsetenv("OVERCOMMIT_CLASS_NAME")
		})

		It("should mutate pod containers and record an event", func() {
			Overcommit(pod, recorder, k8sClient)

			Expect(pod.Spec.Containers[0].Resources.Requests).To(Equal(expectedRequests))

		})
	})
})
