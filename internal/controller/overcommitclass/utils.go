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

// webhookChanged checks if a webhook configuration has changed
func webhookChanged(updated, current interface{}) bool {
	// Type assertion to MutatingWebhook
	updatedWebhook, okUpdated := updated.(admissionv1.MutatingWebhook)
	currentWebhook, okCurrent := current.(admissionv1.MutatingWebhook)

	if !okUpdated || !okCurrent {
		// If we can't cast, assume they're different
		return true
	}

	// Compare webhook name
	if updatedWebhook.Name != currentWebhook.Name {
		return true
	}

	// Compare rules length
	if len(updatedWebhook.Rules) != len(currentWebhook.Rules) {
		return true
	}

	// Compare client config service
	if updatedWebhook.ClientConfig.Service != nil && currentWebhook.ClientConfig.Service != nil {
		if updatedWebhook.ClientConfig.Service.Name != currentWebhook.ClientConfig.Service.Name ||
			updatedWebhook.ClientConfig.Service.Namespace != currentWebhook.ClientConfig.Service.Namespace {
			return true
		}
	} else if (updatedWebhook.ClientConfig.Service == nil) != (currentWebhook.ClientConfig.Service == nil) {
		return true
	}

	// Compare admission review versions
	if len(updatedWebhook.AdmissionReviewVersions) != len(currentWebhook.AdmissionReviewVersions) {
		return true
	}

	for i, version := range updatedWebhook.AdmissionReviewVersions {
		if i >= len(currentWebhook.AdmissionReviewVersions) || version != currentWebhook.AdmissionReviewVersions[i] {
			return true
		}
	}

	// Compare MatchConditions (including CEL expressions for namespace exclusion)
	if len(updatedWebhook.MatchConditions) != len(currentWebhook.MatchConditions) {
		return true
	}

	for i, updatedCondition := range updatedWebhook.MatchConditions {
		if i >= len(currentWebhook.MatchConditions) {
			return true
		}
		currentCondition := currentWebhook.MatchConditions[i]
		if updatedCondition.Name != currentCondition.Name ||
			updatedCondition.Expression != currentCondition.Expression {
			return true
		}
	}

	// If we reach here, they're likely the same
	return false
}
