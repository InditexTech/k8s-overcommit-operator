// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// validRatio reports whether r is a ratio the OvercommitClass CRD accepts.
// cpuOvercommit and memoryOvercommit are validated to the [0.0001, 1] range.
func validRatio(r float64) bool {
	return r >= 0.0001 && r <= 1
}

// FuzzMutateContainers exercises the request rewriting done by the mutating
// webhook: it must never panic and must keep every generated request inside
// [0, limit], for any input the API admits.
func FuzzMutateContainers(f *testing.F) {
	// Seed corpus: CRD boundaries, typical ratios and the no-op ratio of 1.
	f.Add(int64(1000), int64(1<<30), 0.0001, 0.0001, true, true)
	f.Add(int64(1000), int64(1<<30), 1.0, 1.0, true, true)
	f.Add(int64(500), int64(256<<20), 0.5, 0.8, true, false)
	f.Add(int64(0), int64(0), 0.5, 0.5, true, true)
	f.Add(int64(32000), int64(512<<20), 1.0, 0.0001, true, true)
	f.Add(int64(2000), int64(64<<20), 0.5, 0.5, false, true)

	f.Fuzz(func(t *testing.T, cpuMilli, memBytes int64, cpuRatio, memRatio float64, withLimits, withRequests bool) {
		// Stay inside the CRD-validated domain, and keep limits in a range where
		// the float64 -> int64 conversions in mutateContainers cannot overflow.
		if !validRatio(cpuRatio) || !validRatio(memRatio) {
			t.Skip()
		}
		if cpuMilli < 0 || cpuMilli > 1_000_000_000 || memBytes < 0 || memBytes > 1<<50 {
			t.Skip()
		}

		container := corev1.Container{Name: "fuzzed"}
		var cpuLimit, memLimit resource.Quantity
		if withLimits {
			cpuLimit = *resource.NewMilliQuantity(cpuMilli, resource.DecimalSI)
			memLimit = *resource.NewQuantity(memBytes, resource.BinarySI)
			container.Resources.Limits = corev1.ResourceList{
				corev1.ResourceCPU:    cpuLimit,
				corev1.ResourceMemory: memLimit,
			}
		}
		if withRequests {
			container.Resources.Requests = corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewMilliQuantity(1, resource.DecimalSI),
				corev1.ResourceMemory: *resource.NewQuantity(1, resource.BinarySI),
			}
		}

		containers := []corev1.Container{container}
		mutateContainers(containers, cpuRatio, memRatio)

		requests := containers[0].Resources.Requests
		if !withLimits {
			// No limits means no overcommit: requests must be left untouched.
			if len(requests) != len(container.Resources.Requests) {
				t.Fatalf("requests changed without limits: %v", requests)
			}
			return
		}

		if cpuRatio == 1 && memRatio == 1 {
			// A ratio of 1 is a no-op and must not invent request entries.
			if _, ok := requests[corev1.ResourceCPU]; ok && !withRequests {
				t.Fatalf("cpu request created for ratio 1: %v", requests)
			}
			return
		}

		if cpuRatio != 1 {
			cpuReq, ok := requests[corev1.ResourceCPU]
			if !ok {
				t.Fatalf("cpu request missing for ratio %v", cpuRatio)
			}
			if got, limit := cpuReq.MilliValue(), cpuLimit.MilliValue(); got < 0 || got > limit {
				t.Fatalf("cpu request out of bounds: got %d, limit %d, ratio %v", got, limit, cpuRatio)
			}
		}
		if memRatio != 1 {
			memReq, ok := requests[corev1.ResourceMemory]
			if !ok {
				t.Fatalf("memory request missing for ratio %v", memRatio)
			}
			if got, limit := memReq.Value(), memLimit.Value(); got < 0 || got > limit {
				t.Fatalf("memory request out of bounds: got %d, limit %d, ratio %v", got, limit, memRatio)
			}
		}
	})
}
