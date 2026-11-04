// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// validRatio reports whether r is a ratio the OvercommitClass CRD accepts.
// cpuOvercommit and memoryOvercommit are validated to the [0.0001, 1] range.
func validRatio(r float64) bool {
	return r >= 0.0001 && r <= 1
}

// fuzzInput is one generated case for the mutation target.
type fuzzInput struct {
	cpuMilli     int64
	memBytes     int64
	cpuRatio     float64
	memRatio     float64
	withLimits   bool
	withRequests bool
}

// inDomain keeps the target inside the values the API and the CRD admit:
// ratios in [0.0001, 1] and limits small enough that the float64 -> int64
// conversions in mutateContainers cannot overflow.
func (in fuzzInput) inDomain() bool {
	return validRatio(in.cpuRatio) && validRatio(in.memRatio) &&
		in.cpuMilli >= 0 && in.cpuMilli <= 1_000_000_000 &&
		in.memBytes >= 0 && in.memBytes <= 1<<50
}

// container builds the input container and the limits it declares.
func (in fuzzInput) container() (corev1.Container, corev1.ResourceList) {
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewMilliQuantity(in.cpuMilli, resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(in.memBytes, resource.BinarySI),
	}

	c := corev1.Container{Name: "fuzzed"}
	if in.withLimits {
		c.Resources.Limits = limits
	}
	if in.withRequests {
		c.Resources.Requests = corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewMilliQuantity(1, resource.DecimalSI),
			corev1.ResourceMemory: *resource.NewQuantity(1, resource.BinarySI),
		}
	}
	return c, limits
}

// FuzzMutateContainers exercises the request rewriting done by the mutating
// webhook: it must never panic and must keep every generated request inside
// [0, limit], for any input the API admits.
func FuzzMutateContainers(f *testing.F) {
	for _, seed := range []fuzzInput{
		{1000, 1 << 30, 0.0001, 0.0001, true, true},
		{1000, 1 << 30, 1.0, 1.0, true, true},
		{500, 256 << 20, 0.5, 0.8, true, false},
		{0, 0, 0.5, 0.5, true, true},
		{32000, 512 << 20, 1.0, 0.0001, true, true},
		{2000, 64 << 20, 0.5, 0.5, false, true},
	} {
		f.Add(seed.cpuMilli, seed.memBytes, seed.cpuRatio, seed.memRatio, seed.withLimits, seed.withRequests)
	}

	f.Fuzz(func(t *testing.T, cpuMilli, memBytes int64, cpuRatio, memRatio float64, withLimits, withRequests bool) {
		in := fuzzInput{cpuMilli, memBytes, cpuRatio, memRatio, withLimits, withRequests}
		if !in.inDomain() {
			t.Skip()
		}

		original, limits := in.container()
		containers := []corev1.Container{original}
		mutateContainers(containers, in.cpuRatio, in.memRatio)

		assertRequests(t, containers[0].Resources.Requests, original.Resources.Requests, limits, in)
	})
}

// assertRequests checks the invariants of mutateContainers on a single container.
func assertRequests(t *testing.T, got, before, limits corev1.ResourceList, in fuzzInput) {
	t.Helper()

	if !in.withLimits {
		// Without limits there is nothing to overcommit: requests must not move.
		if !reflect.DeepEqual(got, before) {
			t.Fatalf("requests changed without limits: got %v, want %v", got, before)
		}
		return
	}

	assertResource(t, corev1.ResourceCPU, got, before, limits[corev1.ResourceCPU], in.cpuRatio)
	assertResource(t, corev1.ResourceMemory, got, before, limits[corev1.ResourceMemory], in.memRatio)
}

// assertResource checks one resource: a ratio of 1 is a no-op, any other ratio
// produces a request in [0, limit].
func assertResource(t *testing.T, key corev1.ResourceName, got, before corev1.ResourceList, limit resource.Quantity, ratio float64) {
	t.Helper()

	req, present := got[key]
	prev, hadPrev := before[key]

	if ratio == 1 {
		if hadPrev && (!present || prev.Cmp(req) != 0) {
			t.Fatalf("%s request changed for ratio 1: got %v, want %v", key, req.String(), prev.String())
		}
		return
	}

	if !present {
		t.Fatalf("%s request missing for ratio %v", key, ratio)
	}
	if req.Sign() < 0 || req.Cmp(limit) > 0 {
		t.Fatalf("%s request out of bounds: got %s, limit %s, ratio %v", key, req.String(), limit.String(), ratio)
	}
}
