// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"fmt"

	overcommit "github.com/InditexTech/k8s-overcommit-operator/api/v1alphav1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var podlog = logf.Log.WithName("utils")

func GetOvercommitClassSpec(ctx context.Context, name string, k8sClient client.Client) (*overcommit.OvercommitClassSpec, error) {
	// Validate the parameters
	if name == "" {
		return nil, errors.New("name parameter cannot be empty")
	}

	if k8sClient == nil {
		return nil, errors.New("client parameter cannot be nil")
	}

	// Create a new OvercommitClass object
	var overcommitClass overcommit.OvercommitClass

	// Search for the OvercommitClass with the name "cluster"
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name: name,
	}, &overcommitClass)

	if err != nil {
		podlog.Error(err, "Error getting the CR", "name", name)
		return nil, fmt.Errorf("error getting OvercommitClass with name '%s': %w", name, err)
	}

	podlog.Info("OvercommitClass found", "name", name)
	// Return the overcommit label
	return &overcommitClass.Spec, nil
}

func GetDefaultSpec(ctx context.Context, k8sClient client.Client) (*overcommit.OvercommitClassSpec, error) {
	if k8sClient == nil {
		return nil, errors.New("client parameter cannot be nil")
	}

	// List all OvercommitClass
	var overcommitClasses overcommit.OvercommitClassList
	if err := k8sClient.List(ctx, &overcommitClasses); err != nil {
		return nil, fmt.Errorf("error listing OvercommitClass: %w", err)
	}

	// Find the default OvercommitClass
	for _, overcommitClass := range overcommitClasses.Items {
		if overcommitClass.Spec.IsDefault {
			podlog.Info("Default OvercommitClass found", "name", overcommitClass.Name)
			return &overcommitClass.Spec, nil
		}
	}

	return nil, errors.New("no OvercommitClass with isDefault: true found")
}
