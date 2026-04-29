// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

// Package v1alphav1 implements the validating webhook for Pods.
package v1alphav1

import (
	"context"
	"errors"

	"github.com/InditexTech/k8s-overcommit-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	validator := &PodCustomValidator{}
	validator.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr, &corev1.Pod{}).
		WithValidator(validator).
		Complete()
}

// +kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=validating-pod-v1.overcommit.inditex.dev,admissionReviewVersions=v1

// PodCustomValidator struct is responsible for validating the Pod resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type PodCustomValidator struct {
	Client client.Client
}

var _ admission.Validator[*corev1.Pod] = &PodCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateCreate(ctx context.Context, pod *corev1.Pod) (admission.Warnings, error) {

	label, err := utils.GetOvercommitLabel(ctx, v.Client)
	if err != nil {
		return nil, err
	}

	podlog.Info("Validation for Pod upon creation", "name", pod.GetName())

	value, exists := pod.Labels[label]

	if !exists {
		return nil, errors.New("Pod without overcommit class label: " + label)
	}

	unstructuredObj := unstructured.Unstructured{}
	unstructuredObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "overcommit.inditex.dev",
		Version: "v1alphav1",
		Kind:    "OvercommitClass",
	})

	err = v.Client.Get(ctx, client.ObjectKey{
		Name: value,
	}, &unstructuredObj)

	if err != nil {
		return nil, err
	}

	if unstructuredObj.Object == nil {
		return nil, errors.New("OvercommitClass not found")
	}
	return nil, nil

}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateUpdate(ctx context.Context, oldPod *corev1.Pod, pod *corev1.Pod) (admission.Warnings, error) {
	podlog.Info("Validation for Pod upon update", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateDelete(ctx context.Context, pod *corev1.Pod) (admission.Warnings, error) {
	podlog.Info("Validation for Pod upon deletion", "name", pod.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
