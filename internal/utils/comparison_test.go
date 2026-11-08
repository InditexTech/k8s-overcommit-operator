// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTolerationsEqual(t *testing.T) {
	ctx := log.IntoContext(context.Background(), log.Log)

	// Test case 1: Both nil
	if !TolerationsEqual(ctx, nil, nil) {
		t.Error("Expected true for both nil tolerations")
	}

	// Test case 2: One nil, one empty
	empty := []corev1.Toleration{}
	if TolerationsEqual(ctx, nil, empty) {
		t.Error("Expected false for nil vs empty tolerations")
	}

	// Test case 3: Same tolerations
	tolerations1 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	tolerations2 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	if !TolerationsEqual(ctx, tolerations1, tolerations2) {
		t.Error("Expected true for same tolerations")
	}

	// Test case 4: Different tolerations
	tolerations3 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/unreachable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	if TolerationsEqual(ctx, tolerations1, tolerations3) {
		t.Error("Expected false for different tolerations")
	}

	// Test case 5: Different lengths
	tolerations4 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/unreachable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	if TolerationsEqual(ctx, tolerations1, tolerations4) {
		t.Error("Expected false for different lengths")
	}

	// Test case 6: Same tolerations in different order
	tolerations5 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/unreachable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	tolerations6 := []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/unreachable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	if !TolerationsEqual(ctx, tolerations5, tolerations6) {
		t.Error("Expected true for same tolerations in different order")
	}

	// Test case 7: Same tolerations with different TolerationSeconds
	seconds30 := int64(30)
	seconds600 := int64(600)
	tolerations7 := []corev1.Toleration{
		{
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &seconds30,
		},
	}
	tolerations8 := []corev1.Toleration{
		{
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &seconds600,
		},
	}
	if TolerationsEqual(ctx, tolerations7, tolerations8) {
		t.Error("Expected false for different TolerationSeconds")
	}

	// Test case 8: Same TolerationSeconds value behind different pointers
	otherSeconds30 := int64(30)
	tolerations9 := []corev1.Toleration{
		{
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &otherSeconds30,
		},
	}
	if !TolerationsEqual(ctx, tolerations7, tolerations9) {
		t.Error("Expected true for same TolerationSeconds")
	}

	// Test case 9: Zero TolerationSeconds is not the same as unset
	seconds0 := int64(0)
	tolerations10 := []corev1.Toleration{
		{
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &seconds0,
		},
	}
	if TolerationsEqual(ctx, tolerations1, tolerations10) {
		t.Error("Expected false for unset vs zero TolerationSeconds")
	}
}
