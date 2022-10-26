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

	corev1 "k8s.io/api/core/v1"

	crhelperTypes "github.com/pluralsh/controller-reconcile-helper/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WireguardServerSpec defines the desired state of WireguardServer
type WireguardServerSpec struct {
	// Network MTU to use for the VPN
	Mtu string `json:"mtu,omitempty"`

	// +optional
	// Port for the wireguard server
	Port *int32 `json:"port,omitempty"`

	// Service type to use for the VPN
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`

	// +optional
	// ServiceAnnotations for wireguard k8s service
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`

	// WireguardImage for wireguard k8s deployment
	WireguardImage string `json:"wireguardImage"`

	// +optional
	// Sidecars for wireguard k8s deployment
	Sidecars []corev1.Container `json:"sidecars,omitempty"`

	// +optional
	// +kubebuilder:default:="10.8.0.1/24"
	// The CIDR to use for the wireguard server and network
	NetworkCIDR string `json:"networkCIDR,omitempty"`

	// +optional
	// The DNS servers to use for the wireguard server
	DNS []string `json:"dns,omitempty"`

	// The CIDRs that peers can connect to through the wireguard server. Use 0.0.0.0/0 to allow all.
	AllowedIPs []string `json:"allowedIPs,omitempty"`

	// +optional
	// +kubebuilder:default:=false
	// Deploy 3 wireguard servers so that the VPN can be highly available and spread over 3 availability zones
	EnableHA bool `json:"enableHA,omitempty"`

	// +optional
	// The resources to set for the wireguard server
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

const (
	// WireguardServerReadyCondition reports on current status of the Wireguard Server. Ready indicates the instance is in a Running state.
	WireguardServerReadyCondition crhelperTypes.ConditionType = "WireguardServerReady"

	// FailedToCreateServiceReason used when the service could not be created.
	FailedToCreateServiceReason = "FailedToCreateService"

	// FailedToCreateMetricsServiceReason used when the service could not be created.
	FailedToCreateMetricsServiceReason = "FailedToCreateMetricsService"

	// FailedToCreateSecretReason used when the secret could not be created.
	FailedToCreateSecretReason = "FailedToCreateSecret"

	// FailedToCreateConfigMapReason used when the configmap could not be created.
	FailedToCreateConfigMapReason = "FailedToCreateConfigMap"

	// FailedToCreateDeploymentReason used when the configmap could not be created.
	FailedToCreateDeploymentReason = "FailedToCreateDeployment"

	// ServiceNotReadyReason used when service does not yet have a valid IP or hostname
	ServiceNotReadyReason = "ServiceNotReady"

	// InvalidCIDRReason used when the CIDR of the network is invalid
	InvalidCIDRReason = "InvalidCIDR"

	// // InstanceStoppedReason instance is in a stopped state.
	// InstanceStoppedReason = "InstanceStopped"
	// // InstanceNotReadyReason used when the instance is in a pending state.
	// InstanceNotReadyReason = "InstanceNotReady"
	// // InstanceProvisionStartedReason set when the provisioning of an instance started.
	// InstanceProvisionStartedReason = "InstanceProvisionStarted"
	// // InstanceProvisionFailedReason used for failures during instance provisioning.
	// InstanceProvisionFailedReason = "InstanceProvisionFailed"
	// // WaitingForClusterInfrastructureReason used when machine is waiting for cluster infrastructure to be ready before proceeding.
	// WaitingForClusterInfrastructureReason = "WaitingForClusterInfrastructure"
	// // WaitingForBootstrapDataReason used when machine is waiting for bootstrap data to be ready before proceeding.
	// WaitingForBootstrapDataReason = "WaitingForBootstrapData"
)

// WireguardServerStatus defines the observed state of Wireguard
type WireguardServerStatus struct {
	Hostname string `json:"hostname,omitempty"`
	Port     string `json:"port,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Conditions defines current service state of the PacketMachine.
	// +optional
	Conditions crhelperTypes.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Hostname",type="string",JSONPath=".status.hostname",description="WireguardServer hostname"
// +kubebuilder:printcolumn:name="Port",type="string",JSONPath=".status.port",description="WireguardServer port"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="WireguardServer ready status"

// WireguardServer is the Schema for the wireguardservers API
type WireguardServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireguardServerSpec   `json:"spec,omitempty"`
	Status WireguardServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WireguardServerList contains a list of WireguardServer
type WireguardServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WireguardServer `json:"items"`
}

// GetConditions returns the list of conditions for a WireGuardServer API object.
func (ws *WireguardServer) GetConditions() crhelperTypes.Conditions {
	return ws.Status.Conditions
}

// SetConditions will set the given conditions on a WireGuardServer object.
func (ws *WireguardServer) SetConditions(conditions crhelperTypes.Conditions) {
	ws.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&WireguardServer{}, &WireguardServerList{})
}
