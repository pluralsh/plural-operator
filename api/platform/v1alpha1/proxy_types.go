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

// the type for this proxy
// +kubebuilder:validation:Enum=db;sh;web
type ProxyType string

// the rdbms used in this proxy
// +kubebuilder:validation:Enum=postgres;mysql
type EngineType string

const (
	Db ProxyType = "db"
	Sh ProxyType = "sh"
	Web ProxyType = "web"

	Postgres EngineType = "postgres"
	Mysql EngineType = "mysql"
)

// Credentials for authenticating against a proxied resource
type Credentials struct {
	// username to auth with
	User string `json:"user"`
	// secret storing auth info
	Secret string `json:"secret"`
	// key in the secret to use
	Key string `json:"key"`
}

// additional configuration for shell proxies
type ShConfig struct {
	// command to execute in the proxied pod
	Command string `json:"command"`
	// arguments to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`
}

// additional configuration for database proxies
type DbConfig struct {
	// name of the database to connect to
	Name string `json:"name"`
	// db engine
	Engine EngineType `json:"engine"`
	// port to use
	Port int32 `json:"port"`
}

// ProxySpec defines the desired state of Proxy
type ProxySpec struct {
	// Name for this proxy
	Name string `json:"name"`
	// Description for this proxy spec
	// +optional
	Description string `json:"description,omitempty"`
	// the type of proxy to use, can be a db, shell or web proxy
	Type ProxyType `json:"type"`
	// selector to set up the proxy against
	Target string `json:"target"`
	// credentials to use when authenticating against a proxied resource
	// +optional
	Credentials string `json:"credentials"`

	// db-specific configuration for this proxy
	// +optional
	DbConfig *DbConfig `json:"dbConfig,omitempty"`

	// sh-specific configuration for this proxy
	// +optional
	ShConfig *ShConfig `json:"shConfig,omitempty"`
}

// ProxyStatus defines the observed state of Proxy
type ProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Proxy is the Schema for the proxies API
type Proxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxySpec   `json:"spec,omitempty"`
	Status ProxyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxyList contains a list of Proxy
type ProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Proxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Proxy{}, &ProxyList{})
}
