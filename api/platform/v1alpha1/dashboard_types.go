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

// A means of lazily fetching labels for a dashboard
type LabelQuery struct {
	// the backend query to use
	Query string `json:"query"`
	// label name
	Label string `json:"label"`
}

// DashboardLabelSpec is a structure specifying labels to filter against in a dashboard
// these can be statically declared or lazily fetched against the backend metric source
type DashboardLabelSpec struct {
	// label name
	Name string `json:"name"`

	// query to fetch the labels from
	// +optional
	Query *LabelQuery `json:"query,omitempty"`

	// statically specified values
	// +optional
	Values []string `json:"values,omitempty"`
}

// the format for graph x or y values
// +kubebuilder:validation:Enum=bytes;percent;none
type GraphFormat string

const (
	// value is presented in bytes (auto-normalizing to human readable format)
	Bytes GraphFormat = "bytes"
	// value is in percentage points
	Percent GraphFormat = "percent"
	// raw formatting
	None GraphFormat = "none"
)

// Specification for a graph query in a dashboard
type GraphQuery struct {
	// the query to use
	Query string `json:"query"`

	// The format for the legend
	// +optional
	LegendFormat string `json:"legendFormat"`

	// The legend name for this query
	// +optional
	Legend string `json:"legend"`
}

// Specification for a single timeseries graph in a dashboard
type DashboardGraph struct {
	// specify how y values should be rendered. Can be any of [bytes, percent, none]
	// +optional
	Format GraphFormat `json:"format,omitempty"`

	// Name of this graph
	Name string `json:"name"`

	// the queries rendered in this graph
	Queries []*GraphQuery `json:"queries"`
}

// DashboardSpec defines the desired state of Dashboard
type DashboardSpec struct {
	// the name for this dashboard
	Name string `json:"name,omitempty"`
	// description for this dashboard
	Description string `json:"description,omitempty"`

	// possible time windows for the dashboard to display
	Timeslices []string `json:"timeslices"`

	// a list of labels to fetch for filtering dashboard results
	Labels []*DashboardLabelSpec `json:"labels"`

	// the starting time window for dashboard rendering
	DefaultTime string `json:"defaultTime"`

	// the graphs to render in the dashboard
	Graphs []*DashboardGraph `json:"graphs"`
}

// DashboardStatus defines the observed state of Dashboard
type DashboardStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dashboard is the Schema for the dashboards API
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DashboardSpec   `json:"spec,omitempty"`
	Status DashboardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DashboardList contains a list of Dashboard
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dashboard{}, &DashboardList{})
}
