// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

// Package overcommit implements the core overcommit mutation logic.
package overcommit

import (
	"context"

	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getNamespaceOvercommit gets the overcommit values from the namespace label or falls back to the default class.
// Returns (1.0, 1.0) as safe no-op values when any error occurs to avoid mutating pods incorrectly.
func getNamespaceOvercommit(ctx context.Context, pod *corev1.Pod, k8sClient client.Client, label, ownerName, ownerKind string) (float64, float64) {
	// Get the namespace of the pod
	namespaceName := pod.Namespace
	var ns corev1.Namespace
	err := k8sClient.Get(ctx, client.ObjectKey{Name: namespaceName}, &ns)
	if err != nil {
		podlog.Error(err, "Error getting the namespace", "namespace", namespaceName)
		return 1.0, 1.0
	}

	// Check if the overcommit class label is in the namespace
	if val, ok := ns.Labels[label]; ok {
		podlog.Info("Namespace class found", "class", val)
		overcommitClass, err := utils.GetOvercommitClassSpec(ctx, val, k8sClient)
		if err != nil {
			podlog.Error(err, "Error getting the overcommit class", "overcommitClassLabel", val)
			return 1.0, 1.0
		}
		metrics.K8sOvercommitPodMutated.WithLabelValues(val, ownerKind, ownerName, pod.Namespace).Inc()
		return overcommitClass.CpuOvercommit, overcommitClass.MemoryOvercommit
	}

	podlog.Info("Overcommit class not found in the namespace, using the default", "namespace", ns.Name)
	defaultClass, err := utils.GetDefaultSpec(ctx, k8sClient)
	if err != nil {
		podlog.Error(err, "Error getting the default overcommit class")
		return 1.0, 1.0
	}
	metrics.K8sOvercommitPodMutated.WithLabelValues("default", ownerKind, ownerName, pod.Namespace).Inc()
	return defaultClass.CpuOvercommit, defaultClass.MemoryOvercommit
}

func checkOvercommitType(ctx context.Context, pod corev1.Pod, client client.Client) (float64, float64) {
	ownerName, ownerKind, err := utils.GetPodOwner(ctx, client, &pod)
	if err != nil {
		podlog.Error(err, "Error getting the pod owner")
		// Non-fatal: continue with empty owner info
	}

	label, err := utils.GetOvercommitLabel(ctx, client)
	if err != nil {
		podlog.Error(err, "Error getting the overcommit label")
		return 1.0, 1.0
	}
	//  Check if the pod has the overcommit class label
	value, exists := pod.Labels[label]
	podlog.Info(
		"Checking if pod has overcommit class label",
		"overcommitClassLabel", value,
		"exists", exists,
	)
	if exists {
		// Overcommit class found in pod
		overcommitClass, err := utils.GetOvercommitClassSpec(ctx, value, client)
		if err != nil {
			podlog.Error(err, "Error getting the overcommit class", "overcommitClassLabel", value)
			// Overcommit class not found or some error, fall back to namespace/default
			return getNamespaceOvercommit(ctx, &pod, client, label, ownerName, ownerKind)
		}
		metrics.K8sOvercommitPodMutated.WithLabelValues(value, ownerKind, ownerName, pod.Namespace).Inc()
		return overcommitClass.CpuOvercommit, overcommitClass.MemoryOvercommit
	}

	// Overcommit class not found, checking the namespace
	podlog.Info("Overcommit class label not found in pod, checking the namespace")
	return getNamespaceOvercommit(ctx, &pod, client, label, ownerName, ownerKind)
}
