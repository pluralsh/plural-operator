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

package hooks

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-security-plural-sh-v1alpha1-oauthinjector,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=moauthinjector.security.plural.sh,admissionReviewVersions={v1,v1beta1}

type OAuthInjector struct {
	Name          string
	Client        client.Client
	Log           logr.Logger
	decoder       *admission.Decoder
	SidecarConfig *Config
}

type Config struct {
	Containers []corev1.Container `yaml:"containers"`
}

// OidcInjector adds an OAuth2-Proxy sidecar to every incoming pods.
func (oi *OAuthInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := oi.Log.WithValues("webhook", req.AdmissionRequest.Name)
	pod := &corev1.Pod{}

	err := oi.decoder.Decode(req, pod)
	if err != nil {
		log.Info("Sdecar-Injector: cannot decode")
		return admission.Errored(http.StatusBadRequest, err)
	}

	log.Info("Injecting sidecar...")

	secretRef := &[]corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: pod.Annotations["security.plural.sh/oauth-env-secret"],
				},
			},
		},
	}

	sidecarConfig := &Config{}
	sidecarConfig = oi.SidecarConfig
	sidecarConfig.Containers[0].EnvFrom = *secretRef

	pod.Spec.Containers = append(pod.Spec.Containers, sidecarConfig.Containers...)

	log.Info("Sidecar ", oi.Name, " injected.")

	marshaledPod, err := json.Marshal(pod)

	if err != nil {
		log.Info("Sdecar-Injector: cannot marshal")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// OidcInjector implements admission.DecoderInjector.
// A decoder will be automatically inj1ected.
// InjectDecoder injects the decoder.
func (oi *OAuthInjector) InjectDecoder(d *admission.Decoder) error {
	oi.decoder = d
	return nil
}
