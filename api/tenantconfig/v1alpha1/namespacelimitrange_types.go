/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NamespaceLimitRangeSpec defines the desired state of NamespaceLimitRange.
type NamespaceLimitRangeSpec struct {
	Limits []corev1.LimitRangeItem `json:"limits,omitempty"`
}

// NamespaceLimitRangeStatus defines the observed state of NamespaceLimitRange.
type NamespaceLimitRangeStatus struct {
	LimitItems int32 `json:"limitItems,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=nlr
// +kubebuilder:printcolumn:name="LimitItems",type="integer",JSONPath=".status.limitItems"

// NamespaceLimitRange is the Schema for the namespacelimitranges API.
type NamespaceLimitRange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceLimitRangeSpec   `json:"spec,omitempty"`
	Status NamespaceLimitRangeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NamespaceLimitRangeList contains a list of NamespaceLimitRange.
type NamespaceLimitRangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceLimitRange `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespaceLimitRange{}, &NamespaceLimitRangeList{})
}
