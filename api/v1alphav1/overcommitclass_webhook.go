// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var overcommitclasslog = logf.Log.WithName("overcommitclass-resource")

// +kubebuilder:object:generate=false
type OvercommitClassValidator struct {
	// +kubebuilder:skip
	Client client.Client
}

func (v *OvercommitClassValidator) InjectClient(c client.Client) {
	v.Client = c
}

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *OvercommitClass) SetupWebhookWithManager(mgr ctrl.Manager) error {
	validator := &OvercommitClassValidator{}
	validator.InjectClient(mgr.GetClient())
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		WithValidator(validator).
		Complete()
}

// +kubebuilder:webhook:path=/validate-overcommit-inditex-dev-v1alphav1-overcommitclass,mutating=false,failurePolicy=fail,sideEffects=None,groups=overcommit.inditex.dev,resources=overcommitclass,verbs=create;update,versions=v1alphav1,name=overcommitclass.inditex.dev,admissionReviewVersions=v1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (v *OvercommitClassValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {

	overcommitClass, ok := obj.(*OvercommitClass)
	if !ok {
		return nil, fmt.Errorf("failed to cast object to OvercommitClass")
	}
	overcommitclasslog.Info("validate create", "name", overcommitClass.Name)

	err := isClassDefault(*overcommitClass, v.Client)
	if err != nil {
		return nil, err
	}

	err = validateSpecOvercommit(*overcommitClass)
	if err != nil {
		return nil, err
	}

	err = checkDecimals(*overcommitClass)
	if err != nil {
		return nil, err
	}

	err = checkIsRegexValid(overcommitClass.Spec.ExcludedNamespaces)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (v *OvercommitClassValidator) ValidateUpdate(ctx context.Context, old runtime.Object, new runtime.Object) (admission.Warnings, error) {

	// Convert old runtime.Object to *OvercommitClass
	oldOvercommitClass, ok := old.(*OvercommitClass)
	if !ok {
		return nil, fmt.Errorf("failed to cast old object to OvercommitClass")
	}
	overcommitclasslog.Info("validate update", "name", oldOvercommitClass.Name)

	// Convert new runtime.Object to *OvercommitClass
	newOvercommitClass, ok := new.(*OvercommitClass)
	if !ok {
		return nil, fmt.Errorf("failed to cast new object to OvercommitClass")
	}

	if !oldOvercommitClass.Spec.IsDefault {
		err := isClassDefault(*newOvercommitClass, v.Client)
		if err != nil {
			return nil, err
		}
	}

	err := validateSpecOvercommit(*newOvercommitClass)
	if err != nil {
		return nil, err
	}

	err = checkDecimals(*newOvercommitClass)
	if err != nil {
		return nil, err
	}
	err = checkIsRegexValid(newOvercommitClass.Spec.ExcludedNamespaces)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (v *OvercommitClassValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	overcommitClass, ok := obj.(*OvercommitClass)
	if !ok {
		return nil, fmt.Errorf("failed to cast object to OvercommitClass")
	}
	overcommitclasslog.Info("validate delete", "name", overcommitClass.Name)

	return nil, nil
}
