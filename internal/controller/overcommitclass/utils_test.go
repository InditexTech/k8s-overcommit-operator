// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"testing"

	admissionv1 "k8s.io/api/admissionregistration/v1"
)

// The envtest suite in this package needs etcd and kube-apiserver, which are not
// always available (for instance the coverage job that feeds SonarCloud runs a
// plain `go test ./...`). These specs cover webhookChanged without them.
func TestWebhookChangedFailurePolicy(t *testing.T) {
	fail := admissionv1.Fail
	ignore := admissionv1.Ignore

	tests := []struct {
		name     string
		updated  *admissionv1.FailurePolicyType
		current  *admissionv1.FailurePolicyType
		expected bool
	}{
		{
			name:     "Same policy is not a change",
			updated:  &fail,
			current:  &fail,
			expected: false,
		},
		{
			name:     "A different policy is a change",
			updated:  &ignore,
			current:  &fail,
			expected: true,
		},
		{
			name:     "An explicit policy against a missing one is a change",
			updated:  &fail,
			current:  nil,
			expected: true,
		},
		{
			name:     "A missing policy against an explicit one is a change",
			updated:  nil,
			current:  &fail,
			expected: true,
		},
		{
			name:     "Two missing policies are not a change",
			updated:  nil,
			current:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := admissionv1.MutatingWebhook{
				Name:          "economy-overcommit.inditex.dev",
				FailurePolicy: tt.updated,
			}
			current := admissionv1.MutatingWebhook{
				Name:          "economy-overcommit.inditex.dev",
				FailurePolicy: tt.current,
			}

			if changed := webhookChanged(updated, current); changed != tt.expected {
				t.Errorf("webhookChanged() = %v, expected %v", changed, tt.expected)
			}
		})
	}
}
