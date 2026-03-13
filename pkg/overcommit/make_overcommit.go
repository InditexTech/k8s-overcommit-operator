package overcommit

import (
	"context"
	"os"

	"github.com/InditexTech/k8s-overcommit-operator/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var podlog = logf.Log.WithName("overcommit")

func mutateContainers(containers []corev1.Container, pod *corev1.Pod, cpuValue float64, memoryValue float64) {
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

func Overcommit(pod *corev1.Pod, recorder record.EventRecorder, client client.Client) {
	ctx := context.Background()

	metrics.K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	cpuValue, memoryValue := checkOvercommitType(ctx, *pod, client)

	mutateContainers(pod.Spec.Containers, pod, cpuValue, memoryValue)

	// comportamiento actual para CREATE/UPDATE normales
	if len(pod.Spec.InitContainers) > 0 {
		mutateContainers(pod.Spec.InitContainers, pod, cpuValue, memoryValue)
	}

	metrics.K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	recorder.Eventf(
		pod,
		corev1.EventTypeNormal,
		"OvercommitApplied",
		"Applied overcommit to Pod '%s': OvercommitClass = %s, CPU Overcommit = %.2f, Memory Overcommit = %.2f",
		pod.Name,
		os.Getenv("OVERCOMMIT_CLASS_NAME"),
		cpuValue,
		memoryValue,
	)
}

func OvercommitOnResize(pod *corev1.Pod, recorder record.EventRecorder, client client.Client) {
	ctx := context.Background()

	metrics.K8sOvercommitOperatorPodsRequestedTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	cpuValue, memoryValue := checkOvercommitType(ctx, *pod, client)

	// En resize: solo containers normales.
	mutateContainers(pod.Spec.Containers, pod, cpuValue, memoryValue)

	metrics.K8sOvercommitOperatorMutatedPodsTotal.WithLabelValues(os.Getenv("OVERCOMMIT_CLASS_NAME")).Inc()

	recorder.Eventf(
		pod,
		corev1.EventTypeNormal,
		"OvercommitAppliedOnResize",
		"Applied overcommit on resize to Pod '%s': OvercommitClass = %s, CPU Overcommit = %.2f, Memory Overcommit = %.2f",
		pod.Name,
		os.Getenv("OVERCOMMIT_CLASS_NAME"),
		cpuValue,
		memoryValue,
	)
}
