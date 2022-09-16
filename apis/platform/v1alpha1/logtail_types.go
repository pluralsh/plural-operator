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

// LogTailSpec defines the desired state of LogTail
type LogTailSpec struct {
	// the kubectl-type target to use for this log tail, eg deployment/name-of-my-deployment
	Target string `json:"target"`
	// whether to interactively follow the logs
	Follow bool `json:"follow"`
	// number of lines to tail
	Limit int32 `json:"limit"`
	// The specific container to tail
	// +optional
	Container string `json:"container,omitempty"`
}

// LogTailStatus defines the observed state of LogTail
type LogTailStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LogTail is the Schema for the logtails API
type LogTail struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogTailSpec   `json:"spec,omitempty"`
	Status LogTailStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LogTailList contains a list of LogTail
type LogTailList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogTail `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogTail{}, &LogTailList{})
}
