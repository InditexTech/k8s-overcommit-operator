// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

// Package metrics defines Prometheus metrics for the operator.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	K8sOvercommitOperatorPodsRequestedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_overcommit_operator_pods_requested_total",
			Help: "Total number of pods requested to be mutated",
		},
		[]string{"class"},
	)
	K8sOvercommitOperatorMutatedPodsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_overcommit_operator_mutated_pods_total",
			Help: "Total number of pods mutated by the k8s_overcommit_operator webhook",
		},
		[]string{"class"},
	)
	K8sOvercommitOperatorPodsNotMutatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_overcommit_operator_pods_not_mutated_total",
			Help: "Total number of pods not mutated by the k8s_overcommit_operator webhook",
		},
		[]string{"class", "generate_name", "namespace", "reason"},
	)
	K8sOvercommitOperatorTotalClasses = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_overcommit_operator_total_classes",
			Help: "Total number of overcommit classes",
		},
	)
	K8sOvercommitOperatorVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_overcommit_operator_version",
			Help: "K8s_overcommit_operator version",
		},
		[]string{"version"},
	)
	K8sOvercommitOperatorClass = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_overcommit_operator_class",
			Help: "K8s_overcommit_operator class",
		},
		[]string{"name", "cpu", "memory", "isDefault"},
	)
	K8sOvercommitPodMutated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_overcommit_operator_pod_mutated",
			Help: "K8s_overcommit_operator_pod_mutated",
		},
		[]string{"class", "kind", "name", "namespace"},
	)
)

func init() {
	metrics.Registry.MustRegister(K8sOvercommitOperatorPodsRequestedTotal)
	metrics.Registry.MustRegister(K8sOvercommitOperatorMutatedPodsTotal)
	metrics.Registry.MustRegister(K8sOvercommitOperatorPodsNotMutatedTotal)
	metrics.Registry.MustRegister(K8sOvercommitOperatorTotalClasses)
	metrics.Registry.MustRegister(K8sOvercommitOperatorVersion)
	metrics.Registry.MustRegister(K8sOvercommitOperatorClass)
	metrics.Registry.MustRegister(K8sOvercommitPodMutated)
}
