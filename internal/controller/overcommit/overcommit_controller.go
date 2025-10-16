// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	resources "github.com/InditexTech/k8s-overcommit-operator/internal/resources"
	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
)

var logger = logf.Log.WithName("overcommit_controller")

// OvercommitReconciler reconciles a Overcommit object
type OvercommitReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommits/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=overcommit.inditex.dev,resources=overcommits/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Overcommit object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *OvercommitReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger.Info("Starting reconciliation", "name", req.Name, "namespace", req.Namespace, "time", time.Now().Format("15:04:05"))

	label, err := utils.GetOvercommitLabel(ctx, r.Client)
	if err != nil {
		logger.Error(err, "Failed to get Overcommit")
		return ctrl.Result{}, err
	}

	overcommit := &overcommit.Overcommit{}

	err = r.Get(ctx, req.NamespacedName, overcommit)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		// CR not found, nothing to do
		logger.Info("Overcommit CR not found, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	// Migration: Remove legacy finalizers from previous versions
	// Previous versions of this operator added finalizers to CRs for manual cleanup.
	// To ensure smooth upgrades, we automatically remove any legacy finalizers
	// so that CRs created before this change can be deleted normally.
	legacyFinalizers := []string{"overcommit.finalizer", "webhook.finalizer"}
	finalizerRemoved := false
	for _, legacyFinalizer := range legacyFinalizers {
		if controllerutil.ContainsFinalizer(overcommit, legacyFinalizer) {
			logger.Info("Removing legacy finalizer during migration", "finalizer", legacyFinalizer)
			controllerutil.RemoveFinalizer(overcommit, legacyFinalizer)
			finalizerRemoved = true
		}
	}

	// If we removed any legacy finalizers, update the object and requeue
	if finalizerRemoved {
		err = r.Update(ctx, overcommit)
		if err != nil {
			logger.Error(err, "Failed to remove legacy finalizers")
			return ctrl.Result{}, err
		}
		logger.Info("Legacy finalizers removed, requeuing reconciliation")
		return ctrl.Result{RequeueAfter: time.Second * 1}, nil
	}

	// Reconcile Issuer
	issuer := resources.GenerateIssuer()
	if issuer == nil {
		logger.Error(nil, "Generated issuer is nil")
		return ctrl.Result{}, fmt.Errorf("generated issuer is nil")
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, issuer, func() error {
		// Only set controller reference if this is a new resource
		if issuer.CreationTimestamp.IsZero() {
			return ctrl.SetControllerReference(overcommit, issuer, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile issuer")
		return ctrl.Result{}, err
	}

	// Reconcile OvercommitClassValidator
	overcommitClassDeployment := resources.GenerateOvercommitClassValidatingDeployment(*overcommit)
	overcommitClassService := resources.GenerateOvercommitClassValidatingService(*overcommitClassDeployment)
	overcommitClassCertificate := resources.GenerateCertificateValidatingOvercommitClass(*issuer, *overcommitClassService)
	overcommitClassWebhook := resources.GenerateOvercommitClassValidatingWebhookConfiguration(*overcommitClassDeployment, *overcommitClassService, *overcommitClassCertificate)

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, overcommitClassCertificate, func() error {
		// Only set spec if this is a new resource or there are changes
		if overcommitClassCertificate.CreationTimestamp.IsZero() {
			updatedCertificate := resources.GenerateCertificateValidatingOvercommitClass(*issuer, *overcommitClassService)
			overcommitClassCertificate.Spec = updatedCertificate.Spec
			return ctrl.SetControllerReference(overcommit, overcommitClassCertificate, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile OvercommitClass Certificate")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, overcommitClassDeployment, func() error {
		// Regenerate the desired deployment spec
		updatedDeployment := resources.GenerateOvercommitClassValidatingDeployment(*overcommit)
		updatedDeployment.Spec.Template.Spec.Containers[0].Image = os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION")

		// Only update if there are actual differences
		if overcommitClassDeployment.CreationTimestamp.IsZero() {
			// New deployment, set everything
			overcommitClassDeployment.Spec = updatedDeployment.Spec
			overcommitClassDeployment.ObjectMeta.Labels = updatedDeployment.ObjectMeta.Labels
			overcommitClassDeployment.ObjectMeta.Annotations = updatedDeployment.ObjectMeta.Annotations
			return ctrl.SetControllerReference(overcommit, overcommitClassDeployment, r.Scheme)
		} else {
			// Existing deployment, only update specific fields if needed
			updated := false
			if updatedDeployment.Spec.Template.Spec.Containers[0].Image != overcommitClassDeployment.Spec.Template.Spec.Containers[0].Image {
				overcommitClassDeployment.Spec.Template.Spec.Containers[0].Image = updatedDeployment.Spec.Template.Spec.Containers[0].Image
				updated = true
			}
			// Update environment variables if they changed
			if !envVarsEqual(updatedDeployment.Spec.Template.Spec.Containers[0].Env, overcommitClassDeployment.Spec.Template.Spec.Containers[0].Env) {
				overcommitClassDeployment.Spec.Template.Spec.Containers[0].Env = updatedDeployment.Spec.Template.Spec.Containers[0].Env
				updated = true
			}
			// Update template annotations if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Annotations, overcommitClassDeployment.Spec.Template.Annotations) {
				overcommitClassDeployment.Spec.Template.Annotations = updatedDeployment.Spec.Template.Annotations
				updated = true
			}
			// Update template labels if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Labels, overcommitClassDeployment.Spec.Template.Labels) {
				overcommitClassDeployment.Spec.Template.Labels = updatedDeployment.Spec.Template.Labels
				updated = true
			}
			// Update nodeSelector if it changed
			if !mapsEqual(updatedDeployment.Spec.Template.Spec.NodeSelector, overcommitClassDeployment.Spec.Template.Spec.NodeSelector) {
				overcommitClassDeployment.Spec.Template.Spec.NodeSelector = updatedDeployment.Spec.Template.Spec.NodeSelector
				updated = true
			}
			// Update tolerations if they changed
			if !utils.TolerationsEqual(ctx, updatedDeployment.Spec.Template.Spec.Tolerations, overcommitClassDeployment.Spec.Template.Spec.Tolerations) {
				overcommitClassDeployment.Spec.Template.Spec.Tolerations = updatedDeployment.Spec.Template.Spec.Tolerations
				updated = true
			}
			// Only set controller reference if we actually updated something
			if updated {
				return ctrl.SetControllerReference(overcommit, overcommitClassDeployment, r.Scheme)
			}
		}

		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile OvercommitClass Deployment")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, overcommitClassService, func() error {
		// Only update spec if this is a new resource
		if overcommitClassService.CreationTimestamp.IsZero() {
			updatedService := resources.GenerateOvercommitClassValidatingService(*overcommitClassDeployment)
			overcommitClassService.Spec = updatedService.Spec
			return ctrl.SetControllerReference(overcommit, overcommitClassService, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile OvercommitClass Service")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, overcommitClassWebhook, func() error {
		// Only update webhooks if this is a new resource
		if overcommitClassWebhook.CreationTimestamp.IsZero() {
			updatedWebhook := resources.GenerateOvercommitClassValidatingWebhookConfiguration(*overcommitClassDeployment, *overcommitClassService, *overcommitClassCertificate)
			overcommitClassWebhook.Annotations = updatedWebhook.Annotations
			overcommitClassWebhook.Webhooks = updatedWebhook.Webhooks
			return ctrl.SetControllerReference(overcommit, overcommitClassWebhook, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile OvercommitClass Webhook")
		return ctrl.Result{}, err
	}

	// Reconcile PodValidator
	validatingPodDeployment := resources.GeneratePodValidatingDeployment(*overcommit)
	validatingPodService := resources.GeneratePodValidatingService(*validatingPodDeployment)
	validatingpodCertificate := resources.GenerateCertificateValidatingPods(*issuer, *validatingPodService)
	validatingPodWebhook := resources.GeneratePodValidatingWebhookConfiguration(*validatingPodDeployment, *validatingPodService, *validatingpodCertificate, label)

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, validatingpodCertificate, func() error {
		// Only update spec if this is a new resource
		if validatingpodCertificate.CreationTimestamp.IsZero() {
			updatedCertificate := resources.GenerateCertificateValidatingPods(*issuer, *validatingPodService)
			validatingpodCertificate.Spec = updatedCertificate.Spec
			return ctrl.SetControllerReference(overcommit, validatingpodCertificate, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile Pod Validating Certificate")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, validatingPodDeployment, func() error {
		// Regenerate the desired deployment spec
		updatedDeployment := resources.GeneratePodValidatingDeployment(*overcommit)
		updatedDeployment.Spec.Template.Spec.Containers[0].Image = os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION")

		// Only update if there are actual differences
		if validatingPodDeployment.CreationTimestamp.IsZero() {
			// New deployment, set everything
			validatingPodDeployment.Spec = updatedDeployment.Spec
			validatingPodDeployment.ObjectMeta.Labels = updatedDeployment.ObjectMeta.Labels
			validatingPodDeployment.ObjectMeta.Annotations = updatedDeployment.ObjectMeta.Annotations
			return ctrl.SetControllerReference(overcommit, validatingPodDeployment, r.Scheme)
		} else {
			// Existing deployment, only update specific fields if needed
			updated := false
			if updatedDeployment.Spec.Template.Spec.Containers[0].Image != validatingPodDeployment.Spec.Template.Spec.Containers[0].Image {
				validatingPodDeployment.Spec.Template.Spec.Containers[0].Image = updatedDeployment.Spec.Template.Spec.Containers[0].Image
				updated = true
			}
			// Update environment variables if they changed
			if !envVarsEqual(updatedDeployment.Spec.Template.Spec.Containers[0].Env, validatingPodDeployment.Spec.Template.Spec.Containers[0].Env) {
				validatingPodDeployment.Spec.Template.Spec.Containers[0].Env = updatedDeployment.Spec.Template.Spec.Containers[0].Env
				updated = true
			}
			// Update template annotations if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Annotations, validatingPodDeployment.Spec.Template.Annotations) {
				validatingPodDeployment.Spec.Template.Annotations = updatedDeployment.Spec.Template.Annotations
				updated = true
			}
			// Update template labels if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Labels, validatingPodDeployment.Spec.Template.Labels) {
				validatingPodDeployment.Spec.Template.Labels = updatedDeployment.Spec.Template.Labels
				updated = true
			}
			// Update nodeSelector if it changed
			if !mapsEqual(updatedDeployment.Spec.Template.Spec.NodeSelector, validatingPodDeployment.Spec.Template.Spec.NodeSelector) {
				validatingPodDeployment.Spec.Template.Spec.NodeSelector = updatedDeployment.Spec.Template.Spec.NodeSelector
				updated = true
			}
			// Update tolerations if they changed
			if !utils.TolerationsEqual(ctx, updatedDeployment.Spec.Template.Spec.Tolerations, validatingPodDeployment.Spec.Template.Spec.Tolerations) {
				validatingPodDeployment.Spec.Template.Spec.Tolerations = updatedDeployment.Spec.Template.Spec.Tolerations
				updated = true
			}
			// Only set controller reference if we actually updated something
			if updated {
				return ctrl.SetControllerReference(overcommit, validatingPodDeployment, r.Scheme)
			}
		}

		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile Pod Validating Deployment")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, validatingPodService, func() error {
		// Only update spec if this is a new resource
		if validatingPodService.CreationTimestamp.IsZero() {
			updatedService := resources.GeneratePodValidatingService(*validatingPodDeployment)
			validatingPodService.Spec = updatedService.Spec
			return ctrl.SetControllerReference(overcommit, validatingPodService, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile Pod Validating Service")
		return ctrl.Result{}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, validatingPodWebhook, func() error {
		// Only update webhooks if this is a new resource
		if validatingPodWebhook.CreationTimestamp.IsZero() {
			updatedWebhook := resources.GeneratePodValidatingWebhookConfiguration(*validatingPodDeployment, *validatingPodService, *validatingpodCertificate, label)
			validatingPodWebhook.Webhooks = updatedWebhook.Webhooks
			return ctrl.SetControllerReference(overcommit, validatingPodWebhook, r.Scheme)
		}
		return nil
	})
	if err != nil {
		logger.Error(err, "Failed to reconcile Pod Validating Webhook")
		// For resource conflicts, don't fail the reconciliation to avoid immediate retry
		if errors.IsConflict(err) {
			logger.Info("Resource conflict detected, will retry in next scheduled reconciliation")
		} else {
			return ctrl.Result{}, err
		}
	}

	// Reconcile Overcommit Class Controller
	occontroller := resources.GenerateOvercommitClassControllerDeployment(*overcommit)
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, occontroller, func() error {
		// Regenerate the desired deployment spec
		updatedDeployment := resources.GenerateOvercommitClassControllerDeployment(*overcommit)
		updatedDeployment.Spec.Template.Spec.Containers[0].Image = os.Getenv("IMAGE_REGISTRY") + "/" + os.Getenv("IMAGE_REPOSITORY") + ":" + os.Getenv("APP_VERSION")

		// Only update if there are actual differences
		if occontroller.CreationTimestamp.IsZero() {
			// New deployment, set everything
			occontroller.Spec = updatedDeployment.Spec
			occontroller.ObjectMeta.Labels = updatedDeployment.ObjectMeta.Labels
			occontroller.ObjectMeta.Annotations = updatedDeployment.ObjectMeta.Annotations
			logger.Info("Creating new OvercommitClass Controller deployment")
			return ctrl.SetControllerReference(overcommit, occontroller, r.Scheme)
		} else {
			// Existing deployment, only update specific fields if needed
			updated := false
			if updatedDeployment.Spec.Template.Spec.Containers[0].Image != occontroller.Spec.Template.Spec.Containers[0].Image {
				occontroller.Spec.Template.Spec.Containers[0].Image = updatedDeployment.Spec.Template.Spec.Containers[0].Image
				updated = true
			}
			// Update environment variables if they changed
			if !envVarsEqual(updatedDeployment.Spec.Template.Spec.Containers[0].Env, occontroller.Spec.Template.Spec.Containers[0].Env) {
				occontroller.Spec.Template.Spec.Containers[0].Env = updatedDeployment.Spec.Template.Spec.Containers[0].Env
				updated = true
			}
			// Update template annotations if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Annotations, occontroller.Spec.Template.Annotations) {
				occontroller.Spec.Template.Annotations = updatedDeployment.Spec.Template.Annotations
				updated = true
			}
			// Update template labels if they changed
			if !mapsEqual(updatedDeployment.Spec.Template.Labels, occontroller.Spec.Template.Labels) {
				occontroller.Spec.Template.Labels = updatedDeployment.Spec.Template.Labels
				updated = true
			}
			// Update nodeSelector if it changed
			if !mapsEqual(updatedDeployment.Spec.Template.Spec.NodeSelector, occontroller.Spec.Template.Spec.NodeSelector) {
				occontroller.Spec.Template.Spec.NodeSelector = updatedDeployment.Spec.Template.Spec.NodeSelector
				updated = true
			}
			// Update tolerations if they changed
			if !utils.TolerationsEqual(ctx, updatedDeployment.Spec.Template.Spec.Tolerations, occontroller.Spec.Template.Spec.Tolerations) {
				occontroller.Spec.Template.Spec.Tolerations = updatedDeployment.Spec.Template.Spec.Tolerations
				updated = true
			}
			// Only set controller reference if we actually updated something
			if updated {
				return ctrl.SetControllerReference(overcommit, occontroller, r.Scheme)
			}
		}

		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update the status of all resources
	if err := r.updateOvercommitStatusSafely(ctx); err != nil {
		logger.Error(err, "Failed to update Overcommit status")
		// Don't fail the reconciliation for status update errors
	}

	// Only requeue periodically for status checks, not immediately
	logger.Info("Reconciliation completed successfully", "nextReconcile", "10 seconds", "time", time.Now().Format("15:04:05"))
	return ctrl.Result{
		RequeueAfter: time.Second * 10,
	}, nil
}

// +kubebuilder:rbac:groups=apps, resources=deployments;replicasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="", resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io, resources=leases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io, resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

// SetupWithManager sets up the controller with the Manager.
func (r *OvercommitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&overcommit.Overcommit{}).
		Named("Overcommit").
		Complete(r)
}
