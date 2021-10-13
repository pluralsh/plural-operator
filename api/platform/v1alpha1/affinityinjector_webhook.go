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
	"context"
	"strings"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-platform-plural-sh-v1alpha1-affinityinjector,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=maffinityinjector.platform.plural.sh,admissionReviewVersions={v1,v1beta1}

type AffinityInjector struct {
	Name          string
	Client        client.Client
	Log           logr.Logger
	decoder       *admission.Decoder
}

const (
	groupLabel = "platform.plural.sh/resource-groups"
	requiredLabel = "platform.plural.sh/resource-required"
)

// AffinityInjector adds configured node affinities to a pod
func (oi *AffinityInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := oi.Log.WithValues("webhook", req.AdmissionRequest.Name)
	pod := &corev1.Pod{}

	err := oi.decoder.Decode(req, pod)
	if err != nil {
		log.Info("Affinity-Injector: cannot decode")
		return admission.Errored(http.StatusBadRequest, err)
	}

	log.Info("Injecting affinity rules...")

	groupstr, _ := pod.Labels[groupLabel]
	relevantGroup := map[string]bool{}
	for _, group := range strings.Split(groupstr, ",") {
		relevantGroup[group] = true
	}

	var rgs v1alpha1.ResourceGroupList
	if err := oi.Client.List(ctx, &rgs); err != nil {
		log.Error(err, "Failed to list resource groups")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	appliedGroups := make([]*v1alpha1.ResourceGroup, 0)

	for _, group := range rgs.Items {
		if _, ok := relevantGroup[group.Name]; ok {
			appliedGroups = append(appliedGroups, group)
		}
	}

	if len(appliedGroups) == 0 {
		return admission.Allowed("no resource groups applicable")
	}

	affinity := pod.Affinity
	
	if req, ok := pod.Labels[requiredLabel]; ok {
		required := affinity.RequiredDuringSchedulingIgnoredDuringExecution
		terms := required.NodeSelectorTerms
		for _, group := range appliedGroups {
			terms = append(terms, group.Selector)
		}
		affinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: terms,
		}
	} else {
		preferred := affinity.PreferredDuringSchedulingIgnoredDuringExecution
		for _, group := range appliedGroups {
			preferred = append(preferred, &corev1.PreferredSchedulingTerm{
				Weight: 50,
				Preference: group.Selector,
			})
		}
		affinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
	}

	pod.Affinity = affinity
	marshaledPod, err := json.Marshal(pod)

	if err != nil {
		log.Info("Affinity-Injector: cannot marshal")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// AffinityInjector implements admission.DecoderInjector.
// A decoder will be automatically injected.
// InjectDecoder injects the decoder.
func (oi *AffinityInjector) InjectDecoder(d *admission.Decoder) error {
	oi.decoder = d
	return nil
}
