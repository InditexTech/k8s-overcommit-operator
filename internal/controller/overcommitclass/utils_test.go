// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
)

func failurePolicy(policy admissionv1.FailurePolicyType) *admissionv1.FailurePolicyType {
	return &policy
}

// baseWebhook returns a fully populated webhook so that every comparison in
// webhookChanged has something to compare.
func baseWebhook() admissionv1.MutatingWebhook {
	return admissionv1.MutatingWebhook{
		Name: "economy-overcommit.inditex.dev",
		ClientConfig: admissionv1.WebhookClientConfig{
			Service: &admissionv1.ServiceReference{
				Name:      "economy-webhook-service",
				Namespace: "k8s-overcommit",
			},
		},
		Rules:                   []admissionv1.RuleWithOperations{{}},
		AdmissionReviewVersions: []string{"v1"},
		FailurePolicy:           failurePolicy(admissionv1.Fail),
		MatchConditions: []admissionv1.MatchCondition{
			{Name: "exclude-namespaces", Expression: "!object.metadata.namespace.matches('kube-system')"},
		},
	}
}

// webhookChanged is a pure function, so it is covered here and not only in the
// envtest suite, which needs etcd and kube-apiserver to run.
func TestWebhookChanged(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*admissionv1.MutatingWebhook)
		expected bool
	}{
		{
			name:     "An identical webhook is not a change",
			mutate:   func(*admissionv1.MutatingWebhook) {},
			expected: false,
		},
		{
			name:     "A different name is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.Name = "other-overcommit.inditex.dev" },
			expected: true,
		},
		{
			name:     "A different number of rules is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.Rules = nil },
			expected: true,
		},
		{
			name:     "A different service name is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.ClientConfig.Service.Name = "other-webhook-service" },
			expected: true,
		},
		{
			name:     "A different service namespace is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.ClientConfig.Service.Namespace = "other" },
			expected: true,
		},
		{
			name:     "A missing service is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.ClientConfig.Service = nil },
			expected: true,
		},
		{
			name: "A different number of admission review versions is a change",
			mutate: func(w *admissionv1.MutatingWebhook) {
				w.AdmissionReviewVersions = []string{"v1", "v1beta1"}
			},
			expected: true,
		},
		{
			name: "A different admission review version is a change",
			mutate: func(w *admissionv1.MutatingWebhook) {
				w.AdmissionReviewVersions = []string{"v1beta1"}
			},
			expected: true,
		},
		{
			name:     "A different failure policy is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.FailurePolicy = failurePolicy(admissionv1.Ignore) },
			expected: true,
		},
		{
			name:     "A missing failure policy is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.FailurePolicy = nil },
			expected: true,
		},
		{
			name:     "A different match condition name is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.MatchConditions[0].Name = "other" },
			expected: true,
		},
		{
			name:     "A different match condition expression is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.MatchConditions[0].Expression = "true" },
			expected: true,
		},
		{
			name:     "A different number of match conditions is a change",
			mutate:   func(w *admissionv1.MutatingWebhook) { w.MatchConditions = nil },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := baseWebhook()
			current := baseWebhook()
			tt.mutate(&updated)

			if changed := webhookChanged(updated, current); changed != tt.expected {
				t.Errorf("webhookChanged() = %v, expected %v", changed, tt.expected)
			}
		})
	}
}

func TestWebhookChangedWithoutServices(t *testing.T) {
	updated := baseWebhook()
	current := baseWebhook()
	updated.ClientConfig.Service = nil
	current.ClientConfig.Service = nil

	if webhookChanged(updated, current) {
		t.Error("webhookChanged() = true, expected false when neither webhook has a service")
	}
}

func TestWebhookChangedWithUncomparableValues(t *testing.T) {
	if !webhookChanged("not a webhook", "not a webhook") {
		t.Error("webhookChanged() = false, expected true when the values are not MutatingWebhook")
	}
}
