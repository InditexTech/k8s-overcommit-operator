// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"

	resources "github.com/InditexTech/k8s-overcommit-operator/internal/resources"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OvercommitClassReconciler reconciles a OvercommitClass object
type OvercommitClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommitclasses/finalizers,verbs=update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OvercommitClass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile

// SetupWithManager sets up the controller with the Manager.
func (r *OvercommitClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&overcommit.OvercommitClass{}).
		Named("OvercommitClass").
		Complete(r)
}

// cleanupResources ensures that all resources associated with the OvercommitClass CR are deleted.
func (r *OvercommitClassReconciler) cleanupResources(ctx context.Context, overcommitClass *overcommit.OvercommitClass) error {
	logger := log.FromContext(ctx)
	logger.Info("Cleaning up resources associated with OvercommitClass CR", "name", overcommitClass.Name)

	// Delete Deployment
	deployment := resources.CreateDeployment(*overcommitClass)
	if deployment != nil {
		err := r.Delete(ctx, deployment)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete Deployment")
			return err
		}
	}

	// Delete Service
	service := resources.CreateService(overcommitClass.Name)
	if service != nil {
		err := r.Delete(ctx, service)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete Service")
			return err
		}
	}

	// Delete Certificate
	certificate := resources.CreateCertificate(overcommitClass.Name, *service)
	if certificate != nil {
		err := r.Delete(ctx, certificate)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete Certificate")
			return err
		}
	}

	// Delete MutatingWebhookConfiguration
	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit label")
		return err
	}

	webhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)
	if webhookConfig != nil {
		err := r.Delete(ctx, webhookConfig)
		if err != nil && client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete MutatingWebhookConfiguration")
			return err
		}
	}

	logger.Info("Successfully cleaned up resources for OvercommitClass", "name", overcommitClass.Name)
	return nil
}

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

func (r *OvercommitClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting reconciliation", "name", req.Name, "namespace", req.Namespace, "time", time.Now().Format("15:04:05"))

	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit label")
		return ctrl.Result{}, err
	}

	overcommitClass := &overcommit.OvercommitClass{}

	err = r.Get(ctx, req.NamespacedName, overcommitClass)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		// CR not found, nothing to do
		logger.Info("OvercommitClass CR not found, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Check if the CR is being deleted
	if !overcommitClass.ObjectMeta.DeletionTimestamp.IsZero() {
		logger.Info("OvercommitClass CR is being deleted, cleaning up resources")

		// Clean up resources
		err := r.cleanupResources(ctx, overcommitClass)
		if err != nil {
			logger.Error(err, "Failed to clean up resources")
			return ctrl.Result{}, err
		}

		// Remove finalizer if cleanup is successful
		controllerutil.RemoveFinalizer(overcommitClass, "overcommitclass.finalizer")
		err = r.Update(ctx, overcommitClass)
		if err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(overcommitClass, "overcommitclass.finalizer") {
		logger.Info("Adding finalizer to OvercommitClass CR")
		controllerutil.AddFinalizer(overcommitClass, "overcommitclass.finalizer")
		err = r.Update(ctx, overcommitClass)
		if err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		// Return early to trigger a new reconciliation with the updated object
		logger.Info("Finalizer added, requeuing reconciliation")
		return ctrl.Result{}, nil
	}
	// Check if the OvercommitClass has the correct owner reference
	overcommitResource, err := utils.GetOvercommit(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit")
		return ctrl.Result{}, err
	}

	needsOwnerUpdate := false
	if len(overcommitClass.OwnerReferences) == 0 {
		needsOwnerUpdate = true
	} else {
		// Check if the current owner reference is correct
		hasCorrectOwner := false
		for _, ownerRef := range overcommitClass.OwnerReferences {
			if ownerRef.UID == overcommitResource.UID && ownerRef.Kind == "Overcommit" {
				hasCorrectOwner = true
				break
			}
		}
		if !hasCorrectOwner {
			needsOwnerUpdate = true
		}
	}

	if needsOwnerUpdate {
		logger.Info("Setting ControllerReference for OvercommitClass", "name", overcommitClass.Name)
		err = controllerutil.SetControllerReference(&overcommitResource, overcommitClass, r.Scheme)
		if err != nil {
			logger.Error(err, "Failed to set ControllerReference for OvercommitClass")
			return ctrl.Result{}, err
		}

		// Update the OvercommitClass with the new owner reference
		err = r.Update(ctx, overcommitClass)
		if err != nil {
			logger.Error(err, "Failed to update OvercommitClass with ControllerReference")
			return ctrl.Result{}, err
		}
		logger.Info("ControllerReference updated, requeuing reconciliation")
		return ctrl.Result{}, nil
	}

	logger.Info("Reconciling resources for the class", "name", overcommitClass.Name)

	// Create resource definitions
	deployment := resources.CreateDeployment(*overcommitClass)
	service := resources.CreateService(overcommitClass.Name)
	certificate := resources.CreateCertificate(overcommitClass.Name, *service)
	webhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)

	// Reconcile Deployment with improved logic
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// Regenerate the desired deployment spec
		updatedDeployment := resources.CreateDeployment(*overcommitClass)

		// Only update if there are actual differences
		if deployment.CreationTimestamp.IsZero() {
			// New deployment, set everything
			deployment.Spec = updatedDeployment.Spec
			deployment.ObjectMeta.Labels = updatedDeployment.ObjectMeta.Labels
			deployment.ObjectMeta.Annotations = updatedDeployment.ObjectMeta.Annotations
			return controllerutil.SetControllerReference(overcommitClass, deployment, r.Scheme)
		} else {
			// Existing deployment, only update specific fields if needed
			updated := false

			// Check if image changed
			if len(updatedDeployment.Spec.Template.Spec.Containers) > 0 && len(deployment.Spec.Template.Spec.Containers) > 0 {
				if updatedDeployment.Spec.Template.Spec.Containers[0].Image != deployment.Spec.Template.Spec.Containers[0].Image {
					deployment.Spec.Template.Spec.Containers[0].Image = updatedDeployment.Spec.Template.Spec.Containers[0].Image
					updated = true
				}
			}

			// Update environment variables if they changed
			if len(updatedDeployment.Spec.Template.Spec.Containers) > 0 && len(deployment.Spec.Template.Spec.Containers) > 0 {
				if !envVarsEqual(updatedDeployment.Spec.Template.Spec.Containers[0].Env, deployment.Spec.Template.Spec.Containers[0].Env) {
					deployment.Spec.Template.Spec.Containers[0].Env = updatedDeployment.Spec.Template.Spec.Containers[0].Env
					updated = true
				}
			}

			// Update template annotations if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Annotations, deployment.Spec.Template.Annotations) {
				deployment.Spec.Template.Annotations = updatedDeployment.Spec.Template.Annotations
				updated = true
			}

			// Update template labels if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Labels, deployment.Spec.Template.Labels) {
				deployment.Spec.Template.Labels = updatedDeployment.Spec.Template.Labels
				updated = true
			}

			// Only set controller reference if we actually updated something
			if updated {
				return controllerutil.SetControllerReference(overcommitClass, deployment, r.Scheme)
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Deployment")
		return ctrl.Result{}, err
	}

	// Reconcile Service with improved logic
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Only update spec if this is a new resource
		if service.CreationTimestamp.IsZero() {
			updatedService := resources.CreateService(overcommitClass.Name)
			service.Spec = updatedService.Spec
			return controllerutil.SetControllerReference(overcommitClass, service, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Service")
		return ctrl.Result{}, err
	}

	// Reconcile Certificate with improved logic
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, certificate, func() error {
		// Only update spec if this is a new resource
		if certificate.CreationTimestamp.IsZero() {
			updatedCertificate := resources.CreateCertificate(overcommitClass.Name, *service)
			certificate.Spec = updatedCertificate.Spec
			return controllerutil.SetControllerReference(overcommitClass, certificate, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Certificate")
		return ctrl.Result{}, err
	}

	// Reconcile MutatingWebhookConfiguration with improved logic
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, webhookConfig, func() error {
		// Only update webhooks if this is a new resource
		if webhookConfig.CreationTimestamp.IsZero() {
			updatedWebhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)
			webhookConfig.Annotations = updatedWebhookConfig.Annotations
			webhookConfig.Webhooks = updatedWebhookConfig.Webhooks
			return controllerutil.SetControllerReference(overcommitClass, webhookConfig, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update MutatingWebhookConfiguration")
		return ctrl.Result{}, err
	}

	if getTotalClasses(ctx, r.Client) != nil {
		logger.Error(err, "Failed to update metrics")
		return ctrl.Result{}, err
	}

	// Update the status of the resources
	if err := r.updateResourcesStatus(ctx, overcommitClass); err != nil {
		logger.Error(err, "Error updating resource status")
		return ctrl.Result{}, err
	}

	// Only requeue periodically for status checks, not immediately
	logger.Info("Reconciliation completed successfully", "nextReconcile", "10 seconds", "time", time.Now().Format("15:04:05"))
	return ctrl.Result{
		RequeueAfter: 10 * time.Second,
	}, nil
}
