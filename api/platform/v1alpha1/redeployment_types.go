/*
Copyright 2022.

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

// the types of Kubernetes workloads
// +kubebuilder:validation:Enum=daemonsets;deployments;statefulsets
type WorkflowType string

// RedeploymentStatus defines the observed state of Redeployment
type RedeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// RedeploymentSpec is a specification of a redeployment
type RedeploymentSpec struct {

	// Name of the application which must be redeployed after secrets or config maps changes
	Name string `json:"name"`

	// Namespace of the application
	Namespace string `json:"namespace"`
	// The Kubernetes workflow type: DaemonSets, Deployments, StatefulSets
	Workflow WorkflowType `json:"workflow"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Redeployment is the Schema for the redeployment API
type Redeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedeploymentSpec   `json:"spec"`
	Status RedeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedeploymentList contains a list of Redeployment
type RedeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redeployment{}, &RedeploymentList{})
}
