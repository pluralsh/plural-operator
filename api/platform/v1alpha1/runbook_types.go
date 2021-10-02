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

// the type for this datasource
// +kubebuilder:validation:Enum=prometheus;kubernetes;nodes
type DatasourceType string

// the type for this runbook action
// +kubebuilder:validation:Enum=config
type ActionType string

// the type for this kubernetes resource
// +kubebuilder:validation:Enum=deployment;statefulset
type ResourceType string

// the format for a prometheus datasource value
// +kubebuilder:validation:Enum=cpu;memory;none
type FormatType string

const (
	PrometheusDatasourceType DatasourceType = "prometheus"
	KubernetesDatasourceType DatasourceType = "kubernetes"
	NodesDatasourceType      DatasourceType = "nodes"

	ConfigurationActionType ActionType = "config"

	DeploymentResourceType  ResourceType = "deployment"
	StatefulsetResourceType ResourceType = "statefulset"

	CPUFormatType    FormatType = "cpu"
	MemoryFormatType FormatType = "memory"
	NoFormatType     FormatType = "none"
)

// PrometheusDatasource represents a query to prometheus to be used as a runbook datasource
type PrometheusDatasource struct {
	// the prometheus query
	Query string `json:"query"`

	// the format for the value returned
	Format FormatType `json:"format"`

	// the legend to use in the graph of this metric
	Legend string `json:"legend"`
}

// KubernetesDatasource represents a query to the kubernetes api. It only supports individual resources
type KubernetesDatasource struct {
	// the kubernetes resource kind, eg deployment
	Resource ResourceType `json:"resource"`

	// the name of this resource
	Name string `json:"name"`
}

// An update to a configuration path
type PathUpdate struct {
	// path in the configuration to update
	Path []string `json:"path"`

	// the value to use from the args for this execution
	ValueFrom string `json:"valueFrom"`
}

// A representation of a plural configuration update
type ConfigurationAction struct {
	// The updates you want to perform
	Updates []*PathUpdate `json:"updates"`

	// stateful sets to clean before rebuilding (for pvc resizes)
	// +optional
	StatefulSets []string `json:"statefulSets"`
}

// RunbookAction represents an action to be performed in a runbook
type RunbookAction struct {
	// The name to reference this action
	Name string `json:"name"`

	// The type of this action, eg config or kubernetes
	Action ActionType `json:"action"`

	// The url to redirect to after executing this action
	// +optional
	RedirectTo string `json:"redirectTo"`

	// The details of a configuration action
	// +optional
	Configuration *ConfigurationAction `json:"configuration,omitempty"`
}

// RunbookDatasource defines the query to extract data for a runbook
type RunbookDatasource struct {
	// The name to reference this datasource
	Name string `json:"name"`

	// The type of this datasource
	Type DatasourceType `json:"type"`

	// a prometheus query spec
	// +optional
	Prometheus *PrometheusDatasource `json:"prometheus,omitempty"`

	// a kubernetes datasource spec
	// +optional
	Kubernetes *KubernetesDatasource `json:"kubernetes,omitempty"`
}

// RunbookAlert represents an alert to join to this runbook
type RunbookAlert struct {
	// the name of the alert
	Name string `json:"name"`
}

// RunbookAlertStatus represents the status of an alert joined to a runbook
type RunbookAlertStatus struct {
	// the name of the alert
	Name string `json:"name"`

	// the time it fired
	StartsAt string `json:"startsAt"`

	// the alert annotations
	Annotations map[string]string `json:"annotations"`

	// the alert labels
	Labels map[string]string `json:"labels"`

	// the fingerprint of this alert
	Fingerprint string `json:"fingerprint"`
}

// RunbookSpec defines the desired state of Runbook
type RunbookSpec struct {
	// The name for the runbook displayed in the plural console
	Name string `json:"name"`
	// Short description of what this runbook does
	Description string `json:"description"`

	// datasources to hydrate graphs and tables in the runbooks display
	// +optional
	Datasources []*RunbookDatasource `json:"datasources,omitempty"`

	// actions that can be performed in a runbook. These will be references in input forms
	Actions []*RunbookAction `json:"actions"`

	// the display in supported xml for the runbook in the console UI
	Display string `json:"display"`

	// alerts to tie to this runbook
	// +optional
	Alerts []*RunbookAlert `json:"alerts"`
}

// RunbookStatus defines the observed state of Runbook
type RunbookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Alerts []*RunbookAlertStatus `json:"alerts"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Runbook is the Schema for the runbooks API
type Runbook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunbookSpec   `json:"spec,omitempty"`
	Status RunbookStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RunbookList contains a list of Runbook
type RunbookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runbook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Runbook{}, &RunbookList{})
}
