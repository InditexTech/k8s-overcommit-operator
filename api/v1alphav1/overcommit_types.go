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

// OvercommitSpec defines the desired state of Overcommit
type OvercommitSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	OvercommitLabel string `json:"overcommitLabel"`
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// OvercommitStatus defines the observed state of Overcommit
type OvercommitStatus struct {
	Resources  []ResourceStatus   `json:"resources,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Target Label",type=string,JSONPath=".spec.overcommitLabel",description="Label to apply to the pods to make overcommit"
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'cluster'",message="overcommit is a singleton, .metadata.name must be 'cluster'"

// Overcommit is the Schema for the overcommits API
type Overcommit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OvercommitSpec   `json:"spec,omitempty"`
	Status OvercommitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OvercommitList contains a list of Overcommit
type OvercommitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Overcommit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Overcommit{}, &OvercommitList{})
}
