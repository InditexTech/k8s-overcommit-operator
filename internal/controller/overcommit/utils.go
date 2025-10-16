// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	resources "github.com/InditexTech/k8s-overcommit-operator/internal/resources"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

func (r *OvercommitReconciler) updateOvercommitStatus(ctx context.Context, overcommitObject *overcommit.Overcommit) error {
	logger := logf.FromContext(ctx)
	logger.V(1).Info("Updating Overcommit status")

	// Initialize resource status map with better structure
	resourceStatuses := make(map[string]overcommit.ResourceStatus)

	// Helper function to check resource status
	checkResourceStatus := func(name, resourceType string, checkFunc func() error) {
		err := checkFunc()
		ready := err == nil
		resourceStatuses[resourceType] = overcommit.ResourceStatus{
			Name:  name,
			Ready: ready,
		}
		if !ready {
			logger.V(1).Info("Resource not ready", "type", resourceType, "name", name, "error", err)
		}
	}

	// Check Issuer status
	issuer := resources.GenerateIssuer()
	checkResourceStatus(issuer.Name, "issuer", func() error {
		return r.Get(ctx, client.ObjectKey{Name: issuer.Name, Namespace: issuer.Namespace}, issuer)
	})

	// Check OvercommitClass Validator components
	overcommitClassDeployment := resources.GenerateOvercommitClassValidatingDeployment(*overcommitObject)
	checkResourceStatus(overcommitClassDeployment.Name, "overcommitclass-deployment", func() error {
		return r.Get(ctx, client.ObjectKey{Name: overcommitClassDeployment.Name, Namespace: overcommitClassDeployment.Namespace}, overcommitClassDeployment)
	})

	overcommitClassService := resources.GenerateOvercommitClassValidatingService(*overcommitClassDeployment)
	checkResourceStatus(overcommitClassService.Name, "overcommitclass-service", func() error {
		return r.Get(ctx, client.ObjectKey{Name: overcommitClassService.Name, Namespace: overcommitClassService.Namespace}, overcommitClassService)
	})

	overcommitClassCertificate := resources.GenerateCertificateValidatingOvercommitClass(*issuer, *overcommitClassService)
	checkResourceStatus(overcommitClassCertificate.Name, "overcommitclass-certificate", func() error {
		return r.Get(ctx, client.ObjectKey{Name: overcommitClassCertificate.Name, Namespace: overcommitClassCertificate.Namespace}, overcommitClassCertificate)
	})

	overcommitClassWebhook := resources.GenerateOvercommitClassValidatingWebhookConfiguration(*overcommitClassDeployment, *overcommitClassService, *overcommitClassCertificate)
	checkResourceStatus(overcommitClassWebhook.Name, "overcommitclass-webhook", func() error {
		return r.Get(ctx, client.ObjectKey{Name: overcommitClassWebhook.Name}, overcommitClassWebhook)
	})

	// Check Pod Validator components
	podDeployment := resources.GeneratePodValidatingDeployment(*overcommitObject)
	checkResourceStatus(podDeployment.Name, "pod-deployment", func() error {
		return r.Get(ctx, client.ObjectKey{Name: podDeployment.Name, Namespace: podDeployment.Namespace}, podDeployment)
	})

	podService := resources.GeneratePodValidatingService(*podDeployment)
	checkResourceStatus(podService.Name, "pod-service", func() error {
		return r.Get(ctx, client.ObjectKey{Name: podService.Name, Namespace: podService.Namespace}, podService)
	})

	podCertificate := resources.GenerateCertificateValidatingPods(*issuer, *podService)
	checkResourceStatus(podCertificate.Name, "pod-certificate", func() error {
		return r.Get(ctx, client.ObjectKey{Name: podCertificate.Name, Namespace: podCertificate.Namespace}, podCertificate)
	})

	// Check Pod Webhook (handle label errors gracefully)
	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Info("Failed to get Overcommit label, using default", "error", err)
		label = "overcommit.inditex.dev/class"
	}

	podWebhook := resources.GeneratePodValidatingWebhookConfiguration(*podDeployment, *podService, *podCertificate, label)
	checkResourceStatus(podWebhook.Name, "pod-webhook", func() error {
		return r.Get(ctx, client.ObjectKey{Name: podWebhook.Name}, podWebhook)
	})

	// Check OvercommitClass Controller
	ocController := resources.GenerateOvercommitClassControllerDeployment(*overcommitObject)
	checkResourceStatus(ocController.Name, "overcommitclass-controller", func() error {
		return r.Get(ctx, client.ObjectKey{Name: ocController.Name, Namespace: ocController.Namespace}, ocController)
	})

	// Convert map to slice for CRD status (maintain consistent order)
	resourceTypes := []string{
		"issuer",
		"overcommitclass-deployment", "overcommitclass-service", "overcommitclass-certificate", "overcommitclass-webhook",
		"pod-deployment", "pod-service", "pod-certificate", "pod-webhook",
		"overcommitclass-controller",
	}

	resourceStatusSlice := make([]overcommit.ResourceStatus, 0, len(resourceStatuses))
	allReady := true
	readyCount := 0

	for _, resourceType := range resourceTypes {
		if status, exists := resourceStatuses[resourceType]; exists {
			resourceStatusSlice = append(resourceStatusSlice, status)
			if status.Ready {
				readyCount++
			} else {
				allReady = false
			}
		}
	}

	// Update the status of the CRD
	overcommitObject.Status.Resources = resourceStatusSlice

	// Update the condition with more detailed information
	condition := metav1.Condition{
		Type:               "ResourcesReady",
		Status:             metav1.ConditionTrue,
		Reason:             "AllResourcesReady",
		Message:            fmt.Sprintf("All %d managed resources are ready", len(resourceStatusSlice)),
		LastTransitionTime: metav1.Now(),
	}

	if !allReady {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "ResourcesNotReady"
		condition.Message = fmt.Sprintf("%d of %d resources are ready", readyCount, len(resourceStatusSlice))
	}

	setCondition(&overcommitObject.Status, condition)

	// Update the status in the API
	if err := r.Status().Update(ctx, overcommitObject); err != nil {
		logger.Error(err, "Failed to update Overcommit status")
		return err
	}

	logger.V(1).Info("Successfully updated Overcommit status", "ready", readyCount, "total", len(resourceStatusSlice))
	return nil
}

// updateOvercommitStatusSafely safely updates the status by first refreshing the object from the cluster
// with retry logic to handle concurrent modifications
func (r *OvercommitReconciler) updateOvercommitStatusSafely(ctx context.Context) error {
	logger := logf.FromContext(ctx)

	// Since Overcommit is cluster-wide and always named "cluster", use the correct key
	clusterKey := types.NamespacedName{Name: "cluster", Namespace: ""}

	// Retry up to 5 times with exponential backoff
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Fetch the latest version of the object from the cluster
		freshOvercommit := &overcommit.Overcommit{}
		if err := r.Get(ctx, clusterKey, freshOvercommit); err != nil {
			if client.IgnoreNotFound(err) != nil {
				logger.Error(err, "Failed to fetch fresh Overcommit object for status update", "attempt", attempt+1)
				return err
			}
			// Object not found, nothing to update
			logger.V(1).Info("Overcommit object not found, skipping status update")
			return nil
		}

		// Try to update status using the fresh object
		if err := r.updateOvercommitStatus(ctx, freshOvercommit); err != nil {
			isConflict := errors.IsConflict(err)
			isLastAttempt := attempt == maxRetries-1

			if isLastAttempt {
				logger.Error(err, "Failed to update Overcommit status after all retries", "maxRetries", maxRetries)
				return err
			}

			if isConflict {
				// Wait with exponential backoff for conflicts
				backoffDuration := time.Duration(1<<uint(attempt)) * 50 * time.Millisecond
				logger.V(1).Info("Retrying status update due to conflict",
					"attempt", attempt+1,
					"maxRetries", maxRetries,
					"backoff", backoffDuration.String())
				time.Sleep(backoffDuration)
				continue
			} else {
				// Non-conflict error, return immediately
				logger.Error(err, "Non-conflict error during status update")
				return err
			}
		}

		// Success
		logger.V(1).Info("Successfully updated Overcommit status", "attempts", attempt+1)
		return nil
	}

	return fmt.Errorf("failed to update status after %d attempts", maxRetries)
}

func setCondition(status *overcommit.OvercommitStatus, newCondition metav1.Condition) {
	// Ensure LastTransitionTime is set for new conditions
	if newCondition.LastTransitionTime.IsZero() {
		newCondition.LastTransitionTime = metav1.Now()
	}

	// Find existing condition of the same type
	for i, existingCondition := range status.Conditions {
		if existingCondition.Type == newCondition.Type {
			// Check if anything has changed
			if existingCondition.Status != newCondition.Status ||
				existingCondition.Reason != newCondition.Reason ||
				existingCondition.Message != newCondition.Message {

				// Update LastTransitionTime only if status changed
				if existingCondition.Status != newCondition.Status {
					newCondition.LastTransitionTime = metav1.Now()
				} else {
					// Keep the original transition time if only message/reason changed
					newCondition.LastTransitionTime = existingCondition.LastTransitionTime
				}

				status.Conditions[i] = newCondition
			}
			return
		}
	}

	// Condition doesn't exist, add it
	status.Conditions = append(status.Conditions, newCondition)
}

// envVarsEqual compares two slices of environment variables to see if they're equal
// rsEqual compares two slices of environment variables to see if they're equal
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

// annotationsEqual compares two annotation maps to see if they're equal
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
