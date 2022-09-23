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
	"time"

	"sigs.k8s.io/yaml"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/mutate-security-plural-sh-v1alpha1-oauthinjector,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=moauthinjector.security.plural.sh,admissionReviewVersions={v1,v1beta1}

type OAuthInjector struct {
	Name               string
	Client             client.Client
	Log                logr.Logger
	decoder            *admission.Decoder
	ConfigMapName      string
	ConfigMapNamespace string
}

type Config struct {
	Containers []corev1.Container `yaml:"containers"`
}

const (
	htpasswdVolumeName = "htpasswd-secret"
)

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

	kubeconfig := ctrl.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Error(err, "Failed setting up kubernetes client for configmap watcher")
	}

	var sidecarConfig Config

	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, time.Minute, informers.WithNamespace(oi.ConfigMapNamespace))
	configmapInformer := factory.Core().V1().ConfigMaps().Informer()

	configmapInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			if configmap, ok := obj.(*corev1.ConfigMap); ok {
				return configmap.Name == oi.ConfigMapName
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				configmap := obj.(*corev1.ConfigMap)

				containers := configmap.Data["oauth-sidecar-config.yaml"]

				if err := yaml.Unmarshal([]byte(containers), &sidecarConfig); err != nil {
					log.Error(err, "Failed to unmarshal configmap data")
				}

				log.Info("Loaded oauth injector configmap", "config", sidecarConfig)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				newConfigmap := newObj.(*corev1.ConfigMap)

				containers := newConfigmap.Data["oauth-sidecar-config.yaml"]

				if err := yaml.Unmarshal([]byte(containers), &sidecarConfig); err != nil {
					log.Error(err, "Failed to unmarshal configmap data")
				}

				log.Info("Loaded oauth injector configmap", "config", sidecarConfig)
			},
		},
	})

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	sidecarConfig.Containers[0].EnvFrom = *secretRef

	httpwd, ok := pod.Annotations["security.plural.sh/htpasswd-secret"]
	if ok {
		volume := corev1.Volume{Name: htpasswdVolumeName}
		volume.Secret = &corev1.SecretVolumeSource{SecretName: httpwd}
		pod.Spec.Volumes = append(pod.Spec.Volumes, volume)

		sidecarConfig.Containers[0].VolumeMounts = append(sidecarConfig.Containers[0].VolumeMounts, corev1.VolumeMount{
			MountPath: "/etc/plural",
			Name:      htpasswdVolumeName,
		})
	}

	log.Info("Injecting container: ", "container", sidecarConfig.Containers[0])
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
