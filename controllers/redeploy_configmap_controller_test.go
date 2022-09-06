//go:build unit
// +build unit

/*
Copyright 2022.

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

package controllers

import (
	"context"
	"sort"
	"testing"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural-operator/services/redeployment"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
}

func TestReconcileConfigMap(t *testing.T) {
	tests := []struct {
		name            string
		secretName      string
		secretNamespace string
		expectedPods    []string
		existingObjects []ctrlruntimeclient.Object
	}{
		{
			name:            "scenario 1: no matching pods",
			secretNamespace: "test",
			secretName:      "testconfig",
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testconfig",
						Namespace: "test",
					},
					Data: map[string]string{"z": "a", "a": "z"},
				},
				genPodWithConfigMapVolume("pod1", "test", "someconfig"),
			},
			expectedPods: []string{"pod1"},
		},
		{
			name:            "scenario 2: two matching pods, delete those pods",
			secretNamespace: "test",
			secretName:      "testconfig",
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "testconfig",
						Namespace:   "test",
						Annotations: map[string]string{redeployment.ShaAnnotation: "xyz"},
					},
					Data: map[string]string{"z": "a", "a": "z"},
				},
				genPodWithConfigMapVolume("pod1", "test", "someconfig"),
				genPodWithConfigMapVolume("pod2", "test", "testconfig"),
				genPodWithConfigMapVolume("pod3", "test", "someconfig"),
				genPodWithConfigMapVolume("pod4", "test", "testconfig"),
			},
			expectedPods: []string{"pod1", "pod3"},
		},
		{
			name:            "scenario 3: delete pods with config reference and volumes",
			secretNamespace: "test",
			secretName:      "testconfig",
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "testconfig",
						Namespace:   "test",
						Annotations: map[string]string{redeployment.ShaAnnotation: "xyz"},
					},
					Data: map[string]string{"z": "a", "a": "z"},
				},
				genPodWithConfigMapVolume("pod1", "test", "someconfig"),
				genPodWithConfigMapVolume("pod2", "test", "testconfig"),
				genPodWithConfigMapVolume("pod3", "test", "someconfig"),
				genPodWithConfigMapVolume("pod4", "test", "testconfig"),
				genPodWithConfigRef("pod5", "test", "testconfig"),
				genPodWithConfigRef("pod6", "test", "someconfig"),
			},
			expectedPods: []string{"pod1", "pod3", "pod6"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// setup the test scenario
			fakeClient := fake.
				NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(test.existingObjects...).
				Build()

			// act
			ctx := context.Background()
			target := ConfigMapRedeployReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("controllers").WithName("RedeploySecretsController"),
				Scheme: scheme.Scheme,
			}

			_, err := target.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: test.secretName, Namespace: test.secretNamespace}})
			assert.NoError(t, err)

			config := &corev1.ConfigMap{}
			err = fakeClient.Get(ctx, client.ObjectKey{Name: test.secretName, Namespace: test.secretNamespace}, config)
			assert.NoError(t, err)

			_, shaAnnotation := config.Annotations[redeployment.ShaAnnotation]
			assert.True(t, shaAnnotation, "expected SHA annotation")

			existingPods := &corev1.PodList{}
			labelSelector, err := redeployment.RedeployLabelSelector()
			assert.NoError(t, err)
			err = fakeClient.List(ctx, existingPods, &client.ListOptions{Namespace: test.secretNamespace, LabelSelector: labelSelector})
			assert.NoError(t, err)
			existingPodNames := []string{}
			for _, pod := range existingPods.Items {
				existingPodNames = append(existingPodNames, pod.Name)
			}
			sort.Strings(existingPodNames)
			sort.Strings(test.expectedPods)

			assert.Equal(t, existingPodNames, test.expectedPods)
		})
	}
}

func genPodWithConfigMapVolume(podName, podNamespace, configName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
			Labels:    map[string]string{redeployment.RedeployLabel: "true"},
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: configName,
					VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configName,
						},
					},
					},
				},
			},
		},
	}
}

func genPodWithConfigRef(podName, podNamespace, configName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
			Labels:    map[string]string{redeployment.RedeployLabel: "true"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: configName,
								},
							},
						},
					},
				},
			},
		},
	}
}
