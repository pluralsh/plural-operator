/*
Copyright 2021.

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

// ResourceGroupSpec defines the desired state of ResourceGroup
type ResourceGroupSpec struct {
	// the node selector to use for this group
	Selector corev1.NodeSelectorTerm `json:"selector"`
}

// ResourceGroupStatus defines the observed state of ResourceGroup
type ResourceGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ResourceGroup is the Schema for the resourcegroups API
type ResourceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceGroupSpec   `json:"spec,omitempty"`
	Status ResourceGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResourceGroupList contains a list of ResourceGroup
type ResourceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceGroup{}, &ResourceGroupList{})
}
