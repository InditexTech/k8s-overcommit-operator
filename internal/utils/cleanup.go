// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

// Package utils provides shared utility functions for the operator.
package utils

import (
	"context"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteResources(ctx context.Context, k8sClient client.Client) error {
	var nameValidatingPodWebhook = "k8s-overcommit-pod-validating-webhook-webhook"
	var nameValidatingOCWebhook = "k8s-overcommit-class-validating-webhook-webhook"

	// Delete Pod Validating Webhook Configuration
	podWebhook := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: nameValidatingPodWebhook}, podWebhook); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	} else {
		if err := k8sClient.Delete(ctx, podWebhook); err != nil {
			return err
		}
	}

	// Delete Overcommit Class Validating Webhook Configuration
	ocWebhook := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: nameValidatingOCWebhook}, ocWebhook); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	} else {
		if err := k8sClient.Delete(ctx, ocWebhook); err != nil {
			return err
		}
	}

	return nil
}
