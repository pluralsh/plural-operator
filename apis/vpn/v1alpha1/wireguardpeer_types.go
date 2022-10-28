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
	crhelperTypes "github.com/pluralsh/controller-reconcile-helper/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrivateKey struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

type Status struct {
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WireguardPeerSpec defines the desired state of WireguardPeer
type WireguardPeerSpec struct {

	// the name of the active wireguard instance
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=1
	WireguardRef string `json:"wireguardRef"`

	// the IP address of the wireguard peer
	Address string `json:"address,omitempty"`

	// the public key of the wireguard peer
	PublicKey string `json:"publicKey,omitempty"`

	// reference to the secret and key containing the private key of the wireguard peer
	PrivateKeyRef corev1.SecretKeySelector `json:"PrivateKeyRef,omitempty"`
}

const (
	// WireguardPeerReadyCondition reports on current status of the Equinix Metal device. Ready indicates the instance is in a Running state.
	WireguardPeerReadyCondition crhelperTypes.ConditionType = "WireguardPeerReady"

	// WireguardServerNotExistReason used when the Wireguard server of the peer does not exist.
	WireguardServerNotExistReason = "WireguardServerNotExist"

	// WireguardServerNotReadyReason used when the Wireguard server of the peer is not ready.
	WireguardServerNotReadyReason = "WireguardServerNotReady"

	// WaitingForConfigReason used when peer doesn't have a configuration set yet.
	WaitingForConfigReason = "WaitingForConfig"

	// FailedToGetPrivateKeyReason used when the private key can't be found when generating the config secret.
	FailedToGetPrivateKeyReason = "FailedToGetPrivateKey"
)

// WireguardPeerStatus defines the observed state of WireguardPeer
type WireguardPeerStatus struct {
	// The configuration of the wireguard peer without the private key
	Config string `json:"config,omitempty"`

	// Reference to the secret containing the configuration of the wireguard peer
	ConfigRef corev1.SecretKeySelector `json:"configRef,omitempty"`

	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Conditions defines current service state of the PacketMachine.
	// +optional
	Conditions crhelperTypes.Conditions `json:"conditions,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Wireguard Server",type="string",JSONPath=".spec.wireguardRef",description="The Wireguard Server this peer belongs to"
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".spec.address",description="The IP address of this wireguard peer"
// +kubebuilder:printcolumn:name="Config Secret",type="string",JSONPath=".status.configRef.name",description="The name of the secret containing the configuration of the wireguard peer"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="WireguardPeer ready status"

// WireguardPeer is the Schema for the wireguardpeers API
type WireguardPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireguardPeerSpec   `json:"spec,omitempty"`
	Status WireguardPeerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WireguardPeerList contains a list of WireguardPeer
type WireguardPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WireguardPeer `json:"items"`
}

// GetConditions returns the list of conditions for a WireGuardServer API object.
func (ws *WireguardPeer) GetConditions() crhelperTypes.Conditions {
	return ws.Status.Conditions
}

// SetConditions will set the given conditions on a WireGuardServer object.
func (ws *WireguardPeer) SetConditions(conditions crhelperTypes.Conditions) {
	ws.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&WireguardPeer{}, &WireguardPeerList{})
}
