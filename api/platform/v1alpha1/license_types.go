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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LicenseSpec defines the desired state of License
type LicenseSpec struct {
	// the reference to a secret containing your license key
	SecretRef *corev1.SecretKeySelector `json:"secretRef"`
}

// LicenseFeature defines a feature allowed by this license
type LicenseFeature struct {
	// the name of the feature
	Name string `json:"name"`
	// description of the feature
	Description string `json:"description"`
}

// LicensePolicy defines the parameters for a license
type LicensePolicy struct {
	// whether this is on a free plan
	Free bool `json:"free"`
	// the features allowed for this plan
	// +optional
	Features []*LicenseFeature `json:"features"`
	// limits attached to this plan
	// +optional
	Limits map[string]int64 `json:"limits"`
	// the plan you're on
	// +optional
	Plan string `json:"plan"`
}

// LicenseStatus defines the observed state of License
type LicenseStatus struct {
	// the policy this license adheres to
	Policy *LicensePolicy `json:"policy"`

	// additional secrets attached to this license
	// +optional
	Secrets map[string]string `json:"secrets"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// License is the Schema for the licenses API
type License struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LicenseSpec   `json:"spec,omitempty"`
	Status LicenseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LicenseList contains a list of License
type LicenseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []License `json:"items"`
}

func init() {
	SchemeBuilder.Register(&License{}, &LicenseList{})
}
