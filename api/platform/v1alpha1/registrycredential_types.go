package v1alpha1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PasswordSecretRef the password secret reference
type PasswordSecretRef struct {
	// The Secret to select from.
	corev1.LocalObjectReference `json:",inline"`

	// Key for Secret data
	Key string `json:"key"`
}

// RegistryCredentialSpec is a specification of registry credentials
type RegistryCredentialSpec struct {
	// Registry username
	Username string `json:"username"`

	// Registry user email address
	Email string `json:"email"`

	// Registry FQDN
	Server string `json:"server"`

	// The password Secret to select from
	PasswordSecretRef PasswordSecretRef `json:"password"`
}

// RegistryCredentialStatus defines the observed state of RegistryCredential
type RegistryCredentialStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type RegistryCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistryCredentialSpec   `json:"spec"`
	Status RegistryCredentialStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RegistryCredentialList contains a list of RegistryCredential
type RegistryCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryCredential `json:"items"`
}

func (r *RegistryCredential) GetPasswordSecretName() string {
	return fmt.Sprintf("registry-secret-%s", r.Name)
}

func init() {
	SchemeBuilder.Register(&RegistryCredential{}, &RegistryCredentialList{})
}
