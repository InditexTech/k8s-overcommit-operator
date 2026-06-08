// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Overcommit Functions", func() {

	Describe("getNamespaceOvercommit", func() {
		It("should return the correct overcommit values from the namespace", func() {
			resolution := getNamespaceOvercommit(context.TODO(), testPod, k8sClient, "inditex.com/overcommit-class", "ownerName", "ownerKind")
			Expect(resolution.className).To(Equal("test-class"))
			Expect(resolution.cpuValue).To(Equal(0.5))
			Expect(resolution.memoryValue).To(Equal(0.5))
		})
	})

	Describe("checkOvercommitType", func() {
		It("should return the correct overcommit values from the pod", func() {
			resolution := checkOvercommitType(context.TODO(), *testPod, k8sClient)
			Expect(resolution.className).To(Equal("test-class"))
			Expect(resolution.cpuValue).To(Equal(0.5))
			Expect(resolution.memoryValue).To(Equal(0.5))
		})

		It("should fallback to namespace overcommit values if pod label is missing", func() {
			pod := testPod.DeepCopy()
			delete(pod.Labels, "inditex.com/overcommit-class")

			resolution := checkOvercommitType(context.TODO(), *pod, k8sClient)
			Expect(resolution.className).To(Equal("test-class"))
			Expect(resolution.cpuValue).To(Equal(0.5))
			Expect(resolution.memoryValue).To(Equal(0.5))
		})
	})
})
