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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultStorageClassSpec defines the desired state of DefaultStorageClass
type DefaultStorageClassSpec struct {
	Name string `json:"name,omitempty"`
}

// DefaultStorageClassStatus defines the observed state of DefaultStorageClass
type DefaultStorageClassStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// DefaultStorageClass is the Schema for the defaultstorageclasses API
type DefaultStorageClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DefaultStorageClassSpec   `json:"spec,omitempty"`
	Status DefaultStorageClassStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DefaultStorageClassList contains a list of DefaultStorageClass
type DefaultStorageClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DefaultStorageClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DefaultStorageClass{}, &DefaultStorageClassList{})
}
