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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

func (r *OvercommitReconciler) updateOvercommitStatus(ctx context.Context, overcommitObject *overcommit.Overcommit) error {
	logger := logf.FromContext(ctx)

	// Initialize resource status map
	resourceStatuses := make(map[string]overcommit.ResourceStatus)

	// Check Issuer status
	issuer := resources.GenerateIssuer()
	err := r.Get(ctx, client.ObjectKey{Name: issuer.Name, Namespace: issuer.Namespace}, issuer)
	if err == nil {
		resourceStatuses["issuer"] = overcommit.ResourceStatus{Name: issuer.Name, Ready: true}
	} else {
		resourceStatuses["issuer"] = overcommit.ResourceStatus{Name: issuer.Name, Ready: false}
	}

	// Check Deployment Validating Class status
	deployment := resources.GenerateOvercommitClassValidatingDeployment(*overcommitObject)
	err = r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)
	if err == nil {
		resourceStatuses["deployment"] = overcommit.ResourceStatus{Name: deployment.Name, Ready: true}
	} else {
		resourceStatuses["deployment"] = overcommit.ResourceStatus{Name: deployment.Name, Ready: false}
	}

	// Check Service Validating Class status
	service := resources.GenerateOvercommitClassValidatingService(*deployment)
	err = r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, service)
	if err == nil {
		resourceStatuses["service"] = overcommit.ResourceStatus{Name: service.Name, Ready: true}
	} else {
		resourceStatuses["service"] = overcommit.ResourceStatus{Name: service.Name, Ready: false}
	}

	// Check Certificate Validating Class status
	certificate := resources.GenerateCertificateValidatingOvercommitClass(*resources.GenerateIssuer(), *service)
	err = r.Get(ctx, client.ObjectKey{Name: certificate.Name, Namespace: certificate.Namespace}, certificate)
	if err == nil {
		resourceStatuses["certificate"] = overcommit.ResourceStatus{Name: certificate.Name, Ready: true}
	} else {
		resourceStatuses["certificate"] = overcommit.ResourceStatus{Name: certificate.Name, Ready: false}
	}

	// Check Webhook Validating Class status
	webhook := resources.GenerateOvercommitClassValidatingWebhookConfiguration(*deployment, *service, *certificate)
	err = r.Get(ctx, client.ObjectKey{Name: webhook.Name}, webhook)
	if err == nil {
		resourceStatuses["webhook"] = overcommit.ResourceStatus{Name: webhook.Name, Ready: true}
	} else {
		resourceStatuses["webhook"] = overcommit.ResourceStatus{Name: webhook.Name, Ready: false}
	}

	// Check Deployment Validating Pods status
	podDeployment := resources.GeneratePodValidatingDeployment(*overcommitObject)
	err = r.Get(ctx, client.ObjectKey{Name: podDeployment.Name, Namespace: podDeployment.Namespace}, podDeployment)
	if err == nil {
		resourceStatuses["podDeployment"] = overcommit.ResourceStatus{Name: podDeployment.Name, Ready: true}
	} else {
		resourceStatuses["podDeployment"] = overcommit.ResourceStatus{Name: podDeployment.Name, Ready: false}
	}

	// Check Service Validating Pods status
	podService := resources.GeneratePodValidatingService(*podDeployment)
	err = r.Get(ctx, client.ObjectKey{Name: podService.Name, Namespace: podService.Namespace}, podService)
	if err == nil {
		resourceStatuses["podService"] = overcommit.ResourceStatus{Name: podService.Name, Ready: true}
	} else {
		resourceStatuses["podService"] = overcommit.ResourceStatus{Name: podService.Name, Ready: false}
	}

	// Check Certificate Validating Pods status
	podCertificate := resources.GenerateCertificateValidatingPods(*resources.GenerateIssuer(), *podService)
	err = r.Get(ctx, client.ObjectKey{Name: podCertificate.Name, Namespace: podCertificate.Namespace}, podCertificate)
	if err == nil {
		resourceStatuses["podCertificate"] = overcommit.ResourceStatus{Name: podCertificate.Name, Ready: true}
	} else {
		resourceStatuses["podCertificate"] = overcommit.ResourceStatus{Name: podCertificate.Name, Ready: false}
	}

	// Check Webhook Validating Pods status
	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit label")
		return err
	}
	podWebhook := resources.GeneratePodValidatingWebhookConfiguration(*podDeployment, *podService, *podCertificate, label)
	err = r.Get(ctx, client.ObjectKey{Name: podWebhook.Name}, podWebhook)
	if err == nil {
		resourceStatuses["podWebhook"] = overcommit.ResourceStatus{Name: podWebhook.Name, Ready: true}
	} else {
		resourceStatuses["podWebhook"] = overcommit.ResourceStatus{Name: podWebhook.Name, Ready: false}
	}

	// Convert map to slice for CRD status
	resourceStatusSlice := make([]overcommit.ResourceStatus, 0, len(resourceStatuses)) // Pre-allocate slice
	allReady := true
	for _, status := range resourceStatuses {
		resourceStatusSlice = append(resourceStatusSlice, status)
		if !status.Ready {
			allReady = false
		}
	}

	// Update the status of the CRD
	overcommitObject.Status.Resources = resourceStatusSlice

	// Update the condition
	condition := metav1.Condition{
		Type:    "ResourcesReady",
		Status:  metav1.ConditionTrue,
		Reason:  "AllResourcesReady",
		Message: "All managed resources are ready",
	}
	if !allReady {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "ResourcesNotReady"
		condition.Message = "Some resources are not ready"
	}
	setCondition(&overcommitObject.Status, condition)

	// Update the status in the API
	if err := r.Status().Update(ctx, overcommitObject); err != nil {
		logger.Error(err, "Failed to update Overcommit status")
		return err
	}

	return nil
}

// updateOvercommitStatusSafely safely updates the status by first refreshing the object from the cluster
// with retry logic to handle concurrent modifications
// Since Overcommit is cluster-wide and always named "cluster", we use a fixed key
func (r *OvercommitReconciler) updateOvercommitStatusSafely(ctx context.Context) error {
	logger := logf.FromContext(ctx)

	// Since Overcommit is cluster-wide and always named "cluster", use the correct key
	clusterKey := types.NamespacedName{Name: "cluster", Namespace: ""}

	// Retry up to 3 times with exponential backoff
	for attempt := 0; attempt < 3; attempt++ {
		// Fetch the latest version of the object from the cluster
		freshOvercommit := &overcommit.Overcommit{}
		if err := r.Get(ctx, clusterKey, freshOvercommit); err != nil {
			if client.IgnoreNotFound(err) != nil {
				logger.Error(err, "Failed to fetch fresh Overcommit object for status update", "attempt", attempt+1)
				return err
			}
			// Object not found, nothing to update
			return nil
		}

		// Try to update status using the fresh object
		if err := r.updateOvercommitStatus(ctx, freshOvercommit); err != nil {
			if attempt == 2 { // Last attempt
				logger.Error(err, "Failed to update Overcommit status after all retries")
				return err
			}
			logger.Info("Retrying status update due to conflict", "attempt", attempt+1, "error", err.Error())
			// Wait a bit before retrying (exponential backoff)
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
			continue
		}
		// Success
		return nil
	}
	return fmt.Errorf("failed to update status after 3 attempts")
}

func setCondition(status *overcommit.OvercommitStatus, newCondition metav1.Condition) {
	for i, existingCondition := range status.Conditions {
		if existingCondition.Type == newCondition.Type {
			if existingCondition.Status != newCondition.Status ||
				existingCondition.Reason != newCondition.Reason ||
				existingCondition.Message != newCondition.Message {
				newCondition.LastTransitionTime = metav1.Now()
				status.Conditions[i] = newCondition
			}
			return
		}
	}
	newCondition.LastTransitionTime = metav1.Now()
	status.Conditions = append(status.Conditions, newCondition)
}

// cleanupResources ensures that all resources associated with the CR are deleted.
func (r *OvercommitReconciler) cleanupResources(ctx context.Context, overcommitObject *overcommit.Overcommit) error {
	logger := logf.FromContext(ctx)
	logger.Info("Cleaning up resources associated with Overcommit CR")

	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit label")
		return err
	}

	// Delete Issuer
	issuer := resources.GenerateIssuer()
	if issuer != nil {
		err := r.Delete(ctx, issuer)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete Issuer")
			return err
		}
	}

	// Delete OvercommitClassValidator resources
	overcommitClassDeployment := resources.GenerateOvercommitClassValidatingDeployment(*overcommitObject)
	overcommitClassService := resources.GenerateOvercommitClassValidatingService(*overcommitClassDeployment)
	overcommitClassCertificate := resources.GenerateCertificateValidatingOvercommitClass(*issuer, *overcommitClassService)
	overcommitClassWebhook := resources.GenerateOvercommitClassValidatingWebhookConfiguration(*overcommitClassDeployment, *overcommitClassService, *overcommitClassCertificate)

	for _, resource := range []client.Object{overcommitClassDeployment, overcommitClassService, overcommitClassCertificate, overcommitClassWebhook} {
		err := r.Delete(ctx, resource)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("Failed to delete resource: %T", resource))
			return err
		}
	}

	// Delete PodValidator resources
	validatingPodDeployment := resources.GeneratePodValidatingDeployment(*overcommitObject)
	validatingPodService := resources.GeneratePodValidatingService(*validatingPodDeployment)
	validatingpodCertificate := resources.GenerateCertificateValidatingPods(*issuer, *validatingPodService)
	validatingPodWebhook := resources.GeneratePodValidatingWebhookConfiguration(*validatingPodDeployment, *validatingPodService, *validatingpodCertificate, label)

	for _, resource := range []client.Object{validatingPodDeployment, validatingPodService, validatingpodCertificate, validatingPodWebhook} {
		err := r.Delete(ctx, resource)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, fmt.Sprintf("Failed to delete resource: %T", resource))
			return err
		}
	}

	occontroller := resources.GenerateOvercommitClassControllerDeployment(*overcommitObject)
	err = r.Delete(ctx, occontroller)
	if err != nil && client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Failed to delete Overcommit Class Controller")
	}
	return nil
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
