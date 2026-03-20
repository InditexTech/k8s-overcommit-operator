// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPodImageDetails(ctx context.Context, client client.Reader) (string, string, string, error) {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNamespace == "" {
		return "", "", "", fmt.Errorf("POD_NAME or POD_NAMESPACE environment variables are not set")
	}

	pod := &corev1.Pod{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: podNamespace,
	}, pod)
	if err != nil {
		return "", "", "", err
	}

	if len(pod.Spec.Containers) > 0 {
		image := pod.Spec.Containers[0].Image
		// Split the image into registry, image, and tag
		parts := strings.Split(image, "/")
		if len(parts) < 2 {
			return "", "", "", fmt.Errorf("invalid image format: %s", image)
		}

		registry := parts[0]
		imageAndTag := strings.Join(parts[1:], "/")
		imageParts := strings.Split(imageAndTag, ":")
		imageName := imageParts[0]
		tag := "latest" // Default tag if not specified
		if len(imageParts) > 1 {
			tag = imageParts[1]
		}

		return registry, imageName, tag, nil
	}

	return "", "", "", fmt.Errorf("no containers found in pod")
}

func GetPodServiceAccount(ctx context.Context, client client.Reader) (string, error) {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNamespace == "" {
		return "", fmt.Errorf("POD_NAME or POD_NAMESPACE environment variables are not set")
	}

	pod := &corev1.Pod{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: podNamespace,
	}, pod)
	if err != nil {
		return "", err
	}

	return pod.Spec.ServiceAccountName, nil
}

// GetPodDeploymentName retrieves the name of the deployment associated with the current pod.
func GetPodDeploymentName() (string, error) {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	if podName == "" || podNamespace == "" {
		return "", fmt.Errorf("POD_NAME or POD_NAMESPACE environment variables are not set")
	}

	// Create a new Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", fmt.Errorf("unable to create in-cluster config: %w", err)
	}

	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return "", fmt.Errorf("unable to create Kubernetes client: %w", err)
	}

	// Retrieve the pod details
	pod := &corev1.Pod{}
	err = k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      podName,
		Namespace: podNamespace,
	}, pod)
	if err != nil {
		return "", fmt.Errorf("unable to get pod details: %w", err)
	}

	// Find the ReplicaSet owner reference
	var replicaSetName string
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "ReplicaSet" {
			replicaSetName = owner.Name
			break
		}
	}

	if replicaSetName == "" {
		return "", fmt.Errorf("no ReplicaSet owner found for pod")
	}

	// Retrieve the ReplicaSet details
	replicaSet := &appsv1.ReplicaSet{}
	err = k8sClient.Get(context.TODO(), types.NamespacedName{
		Name:      replicaSetName,
		Namespace: podNamespace,
	}, replicaSet)
	if err != nil {
		return "", fmt.Errorf("unable to get ReplicaSet details: %w", err)
	}

	// Find the Deployment owner reference
	for _, owner := range replicaSet.OwnerReferences {
		if owner.Kind == "Deployment" {
			return owner.Name, nil
		}
	}

	return "", fmt.Errorf("no Deployment owner found for ReplicaSet")
}
