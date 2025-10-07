// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// TolerationsEqual compares two slices of tolerations to see if they're equal
func TolerationsEqual(ctx context.Context, a, b []corev1.Toleration) bool {
	logger := logf.FromContext(ctx)

	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		logger.V(1).Info("Tolerations comparison: one slice is nil", "aIsNil", a == nil, "bIsNil", b == nil)
		return false
	}

	if len(a) != len(b) {
		logger.V(1).Info("Tolerations comparison: different lengths", "lenA", len(a), "lenB", len(b))
		return false
	}

	// Create a map for efficient comparison
	tolerationMapA := make(map[string]corev1.Toleration)
	for _, tol := range a {
		key := createTolerationKey(tol)
		tolerationMapA[key] = tol
		logger.V(2).Info("Adding toleration to map A", "key", key, "toleration", tol)
	}

	for _, tol := range b {
		key := createTolerationKey(tol)
		if _, exists := tolerationMapA[key]; !exists {
			logger.V(1).Info("Tolerations comparison: toleration not found", "key", key, "toleration", tol)
			return false
		}
		logger.V(2).Info("Found matching toleration in map A", "key", key, "toleration", tol)
	}

	logger.V(1).Info("Tolerations comparison: all tolerations match")
	return true
}

// createTolerationKey creates a unique key for a toleration for comparison purposes
func createTolerationKey(tol corev1.Toleration) string {
	var parts []string
	parts = append(parts, tol.Key)
	parts = append(parts, string(tol.Operator))
	parts = append(parts, tol.Value)
	parts = append(parts, string(tol.Effect))

	if tol.TolerationSeconds != nil {
		parts = append(parts, "seconds-not-nil")
	} else {
		parts = append(parts, "seconds-nil")
	}

	return strings.Join(parts, "-")
}
