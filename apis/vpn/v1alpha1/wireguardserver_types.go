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

	// // Hostname
	// Hostname string `json:"hostname,omitempty"`

	// +optional
	// Port for the wireguard server
	Port *int32 `json:"port,omitempty"`

	// Service type to use for the VPN
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
}

const (
	// DeviceReadyCondition reports on current status of the Equinix Metal device. Ready indicates the instance is in a Running state.
	WireguardServerReadyCondition crhelperTypes.ConditionType = "WireguardServerReady"

	// FailedToCreateService used when the service could not be created.
	FailedToCreateServiceReason = "FailedToCreateService"

	// FailedToCreateMetricsServiceReason used when the service could not be created.
	FailedToCreateMetricsServiceReason = "FailedToCreateMetricsService"

	// FailedToCreateSecretReason used when the secret could not be created.
	FailedToCreateSecretReason = "FailedToCreateSecret"

	// ServiceNotReadyReason used when service does not yet have a valid IP or hostname
	ServiceNotReadyReason = "ServiceNotReady"

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
