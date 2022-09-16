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

// the type of slash command
// +kubebuilder:validation:Enum=deploy
type SlashCommandType string

const (
	DeployCommand SlashCommandType = "deploy"
)

// SlashCommandSpec a slack-type slash command for use in incident chats
type SlashCommandSpec struct {
	// the slash command to type
	Type SlashCommandType `json:"type,omitempty"`
	// a markdown help doc for this command
	Help string `json:"help"`
}

// SlashCommandStatus defines the observed state of SlashCommand
type SlashCommandStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SlashCommand is the Schema for the slashcommands API
type SlashCommand struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlashCommandSpec   `json:"spec,omitempty"`
	Status SlashCommandStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SlashCommandList contains a list of SlashCommand
type SlashCommandList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlashCommand `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SlashCommand{}, &SlashCommandList{})
}
