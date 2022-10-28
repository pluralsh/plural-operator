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

// the types of configuration this overlay can be applied to
// +kubebuilder:validation:Enum=helm;terraform
type ConfigurationType string

// the types of input values we accept
// +kubebuilder:validation:Enum=string;enum;int;list;bool
type ConfigurationInputType string

// OverlayUpdate defines an update to perform for this update
type OverlayUpdate struct {
	// the path to update with
	Path []string `json:"path"`
}

// ConfigurationOverlaySpec defines the desired state of ConfigurationOverlay
type ConfigurationOverlaySpec struct {
	// Name of the configuration input field
	Name string `json:"name"`

	// Top level folder this overlay should live in, default is "general"
	// +optional
	Folder string `json:"folder"`

	// Subfolder this overlay lives in, default is "all"
	// +optional
	Subfolder string `json:"subfolder"`

	// documentation for the specific field
	Documentation string `json:"documentation"`

	// configuration path to update against
	Updates []OverlayUpdate `json:"updates"`

	// the datatype for the value given to the input field
	// +optional
	InputType ConfigurationInputType `json:"inputType"`

	// the values for enum input types
	// +optional
	InputValues []string `json:"inputValues"`

	// type of configuration value
	// +optional
	Type ConfigurationType `json:"type"`
}

// ConfigurationOverlayStatus defines the observed state of ConfigurationOverlay
type ConfigurationOverlayStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigurationOverlay is the Schema for the configurationoverlays API
type ConfigurationOverlay struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigurationOverlaySpec   `json:"spec,omitempty"`
	Status ConfigurationOverlayStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigurationOverlayList contains a list of ConfigurationOverlay
type ConfigurationOverlayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigurationOverlay `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigurationOverlay{}, &ConfigurationOverlayList{})
}
