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

	// Reconcile Deployment
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

			// Update nodeSelector if it changed
			if !mapsEqual(updatedDeployment.Spec.Template.Spec.NodeSelector, deployment.Spec.Template.Spec.NodeSelector) {
				deployment.Spec.Template.Spec.NodeSelector = updatedDeployment.Spec.Template.Spec.NodeSelector
				updated = true
			}

			// Update tolerations if they changed
			if !utils.TolerationsEqual(ctx, updatedDeployment.Spec.Template.Spec.Tolerations, deployment.Spec.Template.Spec.Tolerations) {
				deployment.Spec.Template.Spec.Tolerations = updatedDeployment.Spec.Template.Spec.Tolerations
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

	// Reconcile Service
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Regenerate the desired service spec
		updatedService := resources.CreateService(overcommitClass.Name)

		// Only update if there are actual differences
		if service.CreationTimestamp.IsZero() {
			// New service, set everything
			service.Spec = updatedService.Spec
			service.ObjectMeta.Labels = updatedService.ObjectMeta.Labels
			service.ObjectMeta.Annotations = updatedService.ObjectMeta.Annotations
			return controllerutil.SetControllerReference(overcommitClass, service, r.Scheme)
		} else {
			// Existing service, only update specific fields if needed
			updated := false

			// Check if selector changed
			if !mapsEqual(updatedService.Spec.Selector, service.Spec.Selector) {
				service.Spec.Selector = updatedService.Spec.Selector
				updated = true
			}

			// Check if ports changed
			if !portsEqual(updatedService.Spec.Ports, service.Spec.Ports) {
				service.Spec.Ports = updatedService.Spec.Ports
				updated = true
			}

			// Check if service type changed
			if updatedService.Spec.Type != service.Spec.Type {
				service.Spec.Type = updatedService.Spec.Type
				updated = true
			}

			// Update annotations if they changed
			if !mapsEqual(updatedService.ObjectMeta.Annotations, service.ObjectMeta.Annotations) {
				service.ObjectMeta.Annotations = updatedService.ObjectMeta.Annotations
				updated = true
			}

			// Update labels if they changed
			if !mapsEqual(updatedService.ObjectMeta.Labels, service.ObjectMeta.Labels) {
				service.ObjectMeta.Labels = updatedService.ObjectMeta.Labels
				updated = true
			}

			// Only set controller reference if we actually updated something
			if updated {
				return controllerutil.SetControllerReference(overcommitClass, service, r.Scheme)
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Service")
		return ctrl.Result{}, err
	}

	// Reconcile Certificate
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, certificate, func() error {
		// Regenerate the desired certificate spec
		updatedCertificate := resources.CreateCertificate(overcommitClass.Name, *service)

		// Only update if there are actual differences
		if certificate.CreationTimestamp.IsZero() {
			// New certificate, set everything
			certificate.Spec = updatedCertificate.Spec
			certificate.ObjectMeta.Labels = updatedCertificate.ObjectMeta.Labels
			certificate.ObjectMeta.Annotations = updatedCertificate.ObjectMeta.Annotations
			return controllerutil.SetControllerReference(overcommitClass, certificate, r.Scheme)
		} else {
			// Existing certificate, only update specific fields if needed
			updated := false

			// Check if DNS names changed
			if !slicesEqual(updatedCertificate.Spec.DNSNames, certificate.Spec.DNSNames) {
				certificate.Spec.DNSNames = updatedCertificate.Spec.DNSNames
				updated = true
			}

			// Check if issuer ref changed
			if updatedCertificate.Spec.IssuerRef.Name != certificate.Spec.IssuerRef.Name ||
				updatedCertificate.Spec.IssuerRef.Kind != certificate.Spec.IssuerRef.Kind {
				certificate.Spec.IssuerRef = updatedCertificate.Spec.IssuerRef
				updated = true
			}

			// Check if secret name changed
			if updatedCertificate.Spec.SecretName != certificate.Spec.SecretName {
				certificate.Spec.SecretName = updatedCertificate.Spec.SecretName
				updated = true
			}

			// Update annotations if they changed
			if !mapsEqual(updatedCertificate.ObjectMeta.Annotations, certificate.ObjectMeta.Annotations) {
				certificate.ObjectMeta.Annotations = updatedCertificate.ObjectMeta.Annotations
				updated = true
			}

			// Update labels if they changed
			if !mapsEqual(updatedCertificate.ObjectMeta.Labels, certificate.ObjectMeta.Labels) {
				certificate.ObjectMeta.Labels = updatedCertificate.ObjectMeta.Labels
				updated = true
			}

			// Only set controller reference if we actually updated something
			if updated {
				return controllerutil.SetControllerReference(overcommitClass, certificate, r.Scheme)
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to create or update Certificate")
		return ctrl.Result{}, err
	}

	// Reconcile MutatingWebhookConfiguration
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, webhookConfig, func() error {
		// Regenerate the desired webhook configuration
		updatedWebhookConfig := resources.CreateMutatingWebhookConfiguration(*overcommitClass, *service, *certificate, label)

		// Only update if there are actual differences
		if webhookConfig.CreationTimestamp.IsZero() {
			// New webhook config, set everything
			webhookConfig.Annotations = updatedWebhookConfig.Annotations
			webhookConfig.Labels = updatedWebhookConfig.Labels
			webhookConfig.Webhooks = updatedWebhookConfig.Webhooks
			return controllerutil.SetControllerReference(overcommitClass, webhookConfig, r.Scheme)
		} else {
			// Existing webhook config, only update specific fields if needed
			updated := false

			// Update annotations if they changed
			if !mapsEqual(updatedWebhookConfig.Annotations, webhookConfig.Annotations) {
				webhookConfig.Annotations = updatedWebhookConfig.Annotations
				updated = true
			}

			// Update labels if they changed
			if !mapsEqual(updatedWebhookConfig.Labels, webhookConfig.Labels) {
				webhookConfig.Labels = updatedWebhookConfig.Labels
				updated = true
			}

			// Check if webhooks changed (simplified comparison)
			if len(updatedWebhookConfig.Webhooks) != len(webhookConfig.Webhooks) {
				webhookConfig.Webhooks = updatedWebhookConfig.Webhooks
				updated = true
			} else {
				// Compare each webhook
				for i, updatedWebhook := range updatedWebhookConfig.Webhooks {
					if i < len(webhookConfig.Webhooks) {
						currentWebhook := webhookConfig.Webhooks[i]
						if webhookChanged(updatedWebhook, currentWebhook) {
							webhookConfig.Webhooks = updatedWebhookConfig.Webhooks
							updated = true
							break
						}
					}
				}
			}

			// Only set controller reference if we actually updated something
			if updated {
				return controllerutil.SetControllerReference(overcommitClass, webhookConfig, r.Scheme)
			}
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
