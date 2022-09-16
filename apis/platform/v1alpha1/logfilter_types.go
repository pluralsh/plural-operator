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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// A label to filter logs against
type LogFilterLabel struct {
	// name of the label
	Name string `json:"name"`
	// value of the label
	Value string `json:"value"`
}

// LogFilterSpec defines the desired state of LogFilter
type LogFilterSpec struct {
	// name for this logfilter
	Name string `json:"name"`
	// description for this logfilter
	Description string `json:"description"`
	// loki query to use for the filter
	// +optional
	Query string `json:"query,omitempty"`
	// labels to query against
	// +optional
	Labels []*LogFilterLabel `json:"labels,omitempty"`
}

// LogFilterStatus defines the observed state of LogFilter
type LogFilterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LogFilter is the Schema for the logfilters API
type LogFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogFilterSpec   `json:"spec,omitempty"`
	Status LogFilterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LogFilterList contains a list of LogFilter
type LogFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogFilter{}, &LogFilterList{})
}
