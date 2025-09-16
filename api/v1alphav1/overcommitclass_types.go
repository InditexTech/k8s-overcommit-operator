// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package v1alphav1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// GetAnnotations implements v1.Object.
func (in *OvercommitClass) GetAnnotations() map[string]string {
	return in.ObjectMeta.Annotations
}

// GetCreationTimestamp implements v1.Object.
func (in *OvercommitClass) GetCreationTimestamp() metav1.Time {
	return in.ObjectMeta.CreationTimestamp
}

// GetDeletionGracePeriodSeconds implements v1.Object.
func (in *OvercommitClass) GetDeletionGracePeriodSeconds() *int64 {
	return in.ObjectMeta.DeletionGracePeriodSeconds
}

// GetDeletionTimestamp implements v1.Object.
func (in *OvercommitClass) GetDeletionTimestamp() *metav1.Time {
	return in.ObjectMeta.DeletionTimestamp
}

// GetFinalizers implements v1.Object.
func (in *OvercommitClass) GetFinalizers() []string {
	return in.ObjectMeta.Finalizers
}

// GetGenerateName implements v1.Object.
func (in *OvercommitClass) GetGenerateName() string {
	return in.ObjectMeta.GenerateName
}

// GetGeneration implements v1.Object.
func (in *OvercommitClass) GetGeneration() int64 {
	return in.ObjectMeta.Generation
}

// GetLabels implements v1.Object.
func (in *OvercommitClass) GetLabels() map[string]string {
	return in.ObjectMeta.Labels
}

// GetManagedFields implements v1.Object.
func (in *OvercommitClass) GetManagedFields() []metav1.ManagedFieldsEntry {
	return in.ObjectMeta.ManagedFields
}

// GetName implements v1.Object.
func (in *OvercommitClass) GetName() string {
	return in.ObjectMeta.Name
}

// GetNamespace implements v1.Object.
func (in *OvercommitClass) GetNamespace() string {
	return in.ObjectMeta.Namespace
}

// GetOwnerReferences implements v1.Object.
func (in *OvercommitClass) GetOwnerReferences() []metav1.OwnerReference {
	return in.ObjectMeta.OwnerReferences
}

// GetResourceVersion implements v1.Object.
func (in *OvercommitClass) GetResourceVersion() string {
	return in.ObjectMeta.ResourceVersion
}

// GetUID implements v1.Object.
func (in *OvercommitClass) GetUID() types.UID {
	return in.ObjectMeta.UID
}

// SetAnnotations implements v1.Object.
func (in *OvercommitClass) SetAnnotations(annotations map[string]string) {
	in.ObjectMeta.Annotations = annotations
}

// SetCreationTimestamp implements v1.Object.
func (in *OvercommitClass) SetCreationTimestamp(timestamp metav1.Time) {
	in.ObjectMeta.CreationTimestamp = timestamp
}

// SetDeletionGracePeriodSeconds implements v1.Object.
func (in *OvercommitClass) SetDeletionGracePeriodSeconds(seconds *int64) {
	in.ObjectMeta.DeletionGracePeriodSeconds = seconds
}

// SetDeletionTimestamp implements v1.Object.
func (in *OvercommitClass) SetDeletionTimestamp(timestamp *metav1.Time) {
	in.ObjectMeta.DeletionTimestamp = timestamp
}

// SetFinalizers implements v1.Object.
func (in *OvercommitClass) SetFinalizers(finalizers []string) {
	in.ObjectMeta.Finalizers = finalizers
}

// SetGenerateName implements v1.Object.
func (in *OvercommitClass) SetGenerateName(name string) {
	in.ObjectMeta.GenerateName = name
}

// SetGeneration implements v1.Object.
func (in *OvercommitClass) SetGeneration(generation int64) {
	in.ObjectMeta.Generation = generation
}

// SetLabels implements v1.Object.
func (in *OvercommitClass) SetLabels(labels map[string]string) {
	in.ObjectMeta.Labels = labels
}

// SetManagedFields implements v1.Object.
func (in *OvercommitClass) SetManagedFields(managedFields []metav1.ManagedFieldsEntry) {
	in.ObjectMeta.ManagedFields = managedFields
}

// SetName implements v1.Object.
func (in *OvercommitClass) SetName(name string) {
	in.ObjectMeta.Name = name
}

// SetNamespace implements v1.Object.
func (in *OvercommitClass) SetNamespace(namespace string) {
	in.ObjectMeta.Namespace = namespace
}

// SetOwnerReferences implements v1.Object.
func (in *OvercommitClass) SetOwnerReferences(references []metav1.OwnerReference) {
	in.ObjectMeta.OwnerReferences = references
}

// SetResourceVersion implements v1.Object.
func (in *OvercommitClass) SetResourceVersion(version string) {
	in.ObjectMeta.ResourceVersion = version
}

// SetUID implements v1.Object.
func (in *OvercommitClass) SetUID(uid types.UID) {
	in.ObjectMeta.UID = uid
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
