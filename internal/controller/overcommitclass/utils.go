// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
)

// envVarsEqual compares two slices of environment variables to see if they're equal
func envVarsEqual(a, b []corev1.EnvVar) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison
	mapA := make(map[string]string)
	mapB := make(map[string]string)

	for _, env := range a {
		mapA[env.Name] = env.Value
	}

	for _, env := range b {
		mapB[env.Name] = env.Value
	}

	// Compare maps
	for key, valueA := range mapA {
		if valueB, exists := mapB[key]; !exists || valueA != valueB {
			return false
		}
	}

	return true
}

// mapsEqual compares two string maps to see if they're equal
func mapsEqual(a, b map[string]string) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		if valueB, exists := b[key]; !exists || valueA != valueB {
			return false
		}
	}

	return true
}

// portsEqual compares two slices of service ports to see if they're equal
func portsEqual(a, b []corev1.ServicePort) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison
	mapA := make(map[string]corev1.ServicePort)
	mapB := make(map[string]corev1.ServicePort)

	for _, port := range a {
		mapA[port.Name] = port
	}

	for _, port := range b {
		mapB[port.Name] = port
	}

	// Compare maps
	for name, portA := range mapA {
		if portB, exists := mapB[name]; !exists ||
			portA.Port != portB.Port ||
			portA.TargetPort != portB.TargetPort ||
			portA.Protocol != portB.Protocol {
			return false
		}
	}

	return true
}

// slicesEqual compares two string slices to see if they're equal
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps for easier comparison (handles order independence)
	mapA := make(map[string]bool)
	mapB := make(map[string]bool)

	for _, item := range a {
		mapA[item] = true
	}

	for _, item := range b {
		mapB[item] = true
	}

	// Compare maps
	for item := range mapA {
		if !mapB[item] {
			return false
		}
	}

	return true
}

// failurePolicyEqual compares two failure policy pointers, treating nil as a
// distinct value from any explicit policy.
func failurePolicyEqual(updated, current *admissionv1.FailurePolicyType) bool {
	if updated == nil || current == nil {
		return updated == current
	}
	return *updated == *current
}

// clientConfigChanged compares the service the webhook points at.
func clientConfigChanged(updated, current admissionv1.WebhookClientConfig) bool {
	if updated.Service == nil || current.Service == nil {
		return (updated.Service == nil) != (current.Service == nil)
	}
	return updated.Service.Name != current.Service.Name ||
		updated.Service.Namespace != current.Service.Namespace
}

// admissionReviewVersionsChanged compares the accepted admission review versions.
func admissionReviewVersionsChanged(updated, current []string) bool {
	if len(updated) != len(current) {
		return true
	}
	for i, version := range updated {
		if version != current[i] {
			return true
		}
	}
	return false
}

// matchConditionsChanged compares the match conditions, including the CEL
// expression used for the namespace exclusion.
func matchConditionsChanged(updated, current []admissionv1.MatchCondition) bool {
	if len(updated) != len(current) {
		return true
	}
	for i, condition := range updated {
		if condition.Name != current[i].Name || condition.Expression != current[i].Expression {
			return true
		}
	}
	return false
}

// webhookChanged checks if a webhook configuration has changed
func webhookChanged(updated, current interface{}) bool {
	// Type assertion to MutatingWebhook
	updatedWebhook, okUpdated := updated.(admissionv1.MutatingWebhook)
	currentWebhook, okCurrent := current.(admissionv1.MutatingWebhook)

	if !okUpdated || !okCurrent {
		// If we can't cast, assume they're different
		return true
	}

	return updatedWebhook.Name != currentWebhook.Name ||
		len(updatedWebhook.Rules) != len(currentWebhook.Rules) ||
		clientConfigChanged(updatedWebhook.ClientConfig, currentWebhook.ClientConfig) ||
		admissionReviewVersionsChanged(updatedWebhook.AdmissionReviewVersions, currentWebhook.AdmissionReviewVersions) ||
		// The failure policy is compared so that editing it on the class reaches the
		// already existing MutatingWebhookConfiguration
		!failurePolicyEqual(updatedWebhook.FailurePolicy, currentWebhook.FailurePolicy) ||
		matchConditionsChanged(updatedWebhook.MatchConditions, currentWebhook.MatchConditions)
}
