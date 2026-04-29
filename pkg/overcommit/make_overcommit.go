// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package overcommit

import (
	"context"
	"fmt"
	"os"

	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// AnnotationOvercommitApplied is set on pods after overcommit mutation to ensure idempotency.
	AnnotationOvercommitApplied = "overcommit.inditex.dev/applied"
)

var podlog = logf.Log.WithName("overcommit")

func mutateContainers(containers []corev1.Container, cpuValue float64, memoryValue float64) {
	for i, container := range containers {
		limits := container.Resources.Limits
		requests := container.Resources.Requests
		if requests == nil {
			requests = corev1.ResourceList{}
		}

		if limits == nil {
			continue
		}

		if cpuLimit, ok := limits[corev1.ResourceCPU]; ok && cpuValue != 1 {
			newCPURequest := float64(cpuLimit.MilliValue()) * cpuValue
			requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(newCPURequest), resource.DecimalSI)
		}

		if memoryLimit, ok := limits[corev1.ResourceMemory]; ok && memoryValue != 1 {
			newMemoryRequest := float64(memoryLimit.Value()) * memoryValue
			requests[corev1.ResourceMemory] = *resource.NewQuantity(int64(newMemoryRequest), resource.BinarySI)
		}

		containers[i].Resources.Requests = requests
	}
}

func Overcommit(ctx context.Context, pod *corev1.Pod, recorder record.EventRecorder, client client.Client) {
	className := os.Getenv("OVERCOMMIT_CLASS_NAME")

	metrics.K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues(className).Inc()

	// Idempotency: skip if this pod was already mutated by this class
	if pod.Annotations != nil {
		if applied, ok := pod.Annotations[AnnotationOvercommitApplied]; ok && applied == className {
			podlog.Info("Pod already mutated by this overcommit class, skipping", "pod", pod.Name, "class", className)
			return
		}
	}

	cpuValue, memoryValue := checkOvercommitType(ctx, *pod, client)

	mutateContainers(pod.Spec.Containers, cpuValue, memoryValue)

	// Also mutate init containers on regular CREATE/UPDATE
	if len(pod.Spec.InitContainers) > 0 {
		mutateContainers(pod.Spec.InitContainers, cpuValue, memoryValue)
	}

	// Mark the pod as mutated to prevent double-application on reinvocation
	setOvercommitAnnotation(pod, className, cpuValue, memoryValue)

	metrics.K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues(className).Inc()

	recorder.Eventf(
		pod,
		corev1.EventTypeNormal,
		"OvercommitApplied",
		"Applied overcommit to Pod '%s': OvercommitClass = %s, CPU Overcommit = %.2f, Memory Overcommit = %.2f",
		pod.Name,
		className,
		cpuValue,
		memoryValue,
	)
}

func OvercommitOnResize(ctx context.Context, pod *corev1.Pod, recorder record.EventRecorder, client client.Client) {
	className := os.Getenv("OVERCOMMIT_CLASS_NAME")

	metrics.K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues(className).Inc()

	cpuValue, memoryValue := checkOvercommitType(ctx, *pod, client)

	// On resize: only mutate regular containers, skip init containers.
	mutateContainers(pod.Spec.Containers, cpuValue, memoryValue)

	// Update annotation with new values after resize
	setOvercommitAnnotation(pod, className, cpuValue, memoryValue)

	metrics.K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues(className).Inc()

	recorder.Eventf(
		pod,
		corev1.EventTypeNormal,
		"OvercommitAppliedOnResize",
		"Applied overcommit on resize to Pod '%s': OvercommitClass = %s, CPU Overcommit = %.2f, Memory Overcommit = %.2f",
		pod.Name,
		className,
		cpuValue,
		memoryValue,
	)
}

// setOvercommitAnnotation marks the pod as having been mutated by the overcommit webhook.
func setOvercommitAnnotation(pod *corev1.Pod, className string, cpuValue, memoryValue float64) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[AnnotationOvercommitApplied] = className
	pod.Annotations["overcommit.inditex.dev/cpu"] = fmt.Sprintf("%.4f", cpuValue)
	pod.Annotations["overcommit.inditex.dev/memory"] = fmt.Sprintf("%.4f", memoryValue)
}
