// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OvercommitClassSpec defines the desired state of OvercommitClass
type OvercommitClassSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum=0.0001
	// +kubebuilder:validation:Maximum=1
	// +kubebuilder:validation:Required
	CpuOvercommit float64 `json:"cpuOvercommit,omitempty"`
	// +kubebuilder:validation:Minimum=0.0001
	// +kubebuilder:validation:Maximum=1
	// +kubebuilder:validation:Required
	MemoryOvercommit float64 `json:"memoryOvercommit,omitempty"`
	// +kubebuilder:validation:Required
	ExcludedNamespaces string `json:"excludedNamespaces,omitempty"`
	// +kubebuilder:default=false
	IsDefault   bool              `json:"isDefault,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

type ResourceStatus struct {
	Name  string `json:"name,omitempty"`
	Ready bool   `json:"ready"`
}

// OvercommitClassStatus defines the observed state of OvercommitClass
type OvercommitClassStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Resources  []ResourceStatus   `json:"resources,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=oc;ocs
// +kubebuilder:printcolumn:name="CPU",type=number,JSONPath=".spec.cpuOvercommit",description="CPU overcommit ratio"
// +kubebuilder:printcolumn:name="Memory",type=number,JSONPath=".spec.memoryOvercommit",description="Memory overcommit ratio"
// +kubebuilder:printcolumn:name="Default",type=boolean,JSONPath=".spec.isDefault",description="Is default overcommit class"

// OvercommitClass is the Schema for the overcommitclasses API
type OvercommitClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OvercommitClassSpec   `json:"spec,omitempty"`
	Status OvercommitClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OvercommitClassList contains a list of OvercommitClass
type OvercommitClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OvercommitClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OvercommitClass{}, &OvercommitClassList{})
}
