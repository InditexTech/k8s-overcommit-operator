// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	"context"
	"fmt"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/pkg/overcommit"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.

// PodCustomDefaulter struct is responsible for setting default values on the custom resource of the Kind Pod.
type PodCustomDefaulter struct {
	Recorder record.EventRecorder
	Client   client.Client
}

func (d *PodCustomDefaulter) InjectRecorder(r record.EventRecorder) {
	d.Recorder = r
}

func (d *PodCustomDefaulter) InjectClient(c client.Client) {
	d.Client = c
}

var _ webhook.CustomDefaulter = &PodCustomDefaulter{}

func (d *PodCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected a Pod object but got %T", obj)
	}

	isResize := false
	if req, err := admission.RequestFromContext(ctx); err == nil {
		isResize = req.SubResource == "resize"
	}

	if isResize {
		overcommit.OvercommitOnResize(pod, d.Recorder, d.Client)
		return nil
	}

	overcommit.Overcommit(pod, d.Recorder, d.Client)
	return nil
}

// +kubebuilder:webhook:path=/mutate--v1-pod,mutating=true,failurePolicy=ignore,reinvocationPolicy=IfNeeded,sideEffects=None,groups="",resources=pods;pods/resize,verbs=create;update,versions=v1,name=mutating-pod-v1.overcommit.inditex.dev,admissionReviewVersions=v1
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	defaulter := &PodCustomDefaulter{}
	defaulter.InjectRecorder(mgr.GetEventRecorderFor("pod-defaulter"))
	defaulter.InjectClient(mgr.GetClient())
	return ctrl.NewWebhookManagedBy(mgr).
		For(&corev1.Pod{}).
		WithDefaulter(defaulter).
		Complete()
}
