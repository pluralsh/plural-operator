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
	"strings"

	"github.com/go-logr/logr"
	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-platform-plural-sh-v1alpha1-affinityinjector,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=maffinityinjector.platform.plural.sh,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:rbac:groups="platform.plural.sh",resources=resourcegroups,verbs=get;list;watch

type AffinityInjector struct {
	client.Client
	Name    string
	Log     logr.Logger
	decoder *admission.Decoder
}

const (
	groupLabel    = "platform.plural.sh/resource-groups"
	requiredLabel = "platform.plural.sh/resource-required"
)

// AffinityInjector adds configured node affinities to a pod
func (ai *AffinityInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := ai.Log.WithValues("webhook", req.AdmissionRequest.Name)
	pod := &corev1.Pod{}

	err := ai.decoder.Decode(req, pod)
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

	var rgs platformv1alpha1.ResourceGroupList
	if err := ai.Client.List(ctx, &rgs); err != nil {
		log.Error(err, "Failed to list resource groups")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	appliedGroups := make([]*platformv1alpha1.ResourceGroup, 0)

	for _, group := range rgs.Items {
		if _, ok := relevantGroup[group.Name]; ok {
			appliedGroups = append(appliedGroups, &group)
		}
	}

	if len(appliedGroups) == 0 {
		return admission.Allowed("no resource groups applicable")
	}

	affinity := pod.Spec.Affinity
	if affinity != nil {
		affinity = &corev1.Affinity{}
	}
	nodeAffinity := affinity.NodeAffinity

	if _, ok := pod.Labels[requiredLabel]; ok {
		required := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		terms := required.NodeSelectorTerms
		for _, group := range appliedGroups {
			terms = append(terms, group.Spec.Selector)
		}
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: terms,
		}
	} else {
		preferred := nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		for _, group := range appliedGroups {
			preferred = append(preferred, corev1.PreferredSchedulingTerm{
				Weight:     50,
				Preference: group.Spec.Selector,
			})
		}
		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferred
	}

	affinity.NodeAffinity = nodeAffinity
	pod.Spec.Affinity = affinity
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
func (ai *AffinityInjector) InjectDecoder(d *admission.Decoder) error {
	ai.decoder = d
	return nil
}
