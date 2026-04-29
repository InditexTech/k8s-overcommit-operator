// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

// Package controller implements the OvercommitClass reconciler.
package controller

import (
	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setCondition updates or adds a condition in the status following Kubernetes standards.
func setCondition(status *overcommit.OvercommitClassStatus, newCondition metav1.Condition) {
	for i, existingCondition := range status.Conditions {
		if existingCondition.Type == newCondition.Type {
			// Update only if there are changes
			if existingCondition.Status != newCondition.Status ||
				existingCondition.Reason != newCondition.Reason ||
				existingCondition.Message != newCondition.Message {
				newCondition.LastTransitionTime = metav1.Now()
				status.Conditions[i] = newCondition
			}
			return
		}
	}
	// Add the new condition if it doesn't exist
	newCondition.LastTransitionTime = metav1.Now()
	status.Conditions = append(status.Conditions, newCondition)
}
