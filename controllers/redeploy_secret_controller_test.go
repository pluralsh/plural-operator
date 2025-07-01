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

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"github.com/pluralsh/plural-operator/services/redeployment"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
}

func TestReconcileSecret(t *testing.T) {
	tests := []struct {
		name                  string
		secretName            string
		secretNamespace       string
		expectedPods          []string
		existingObjects       []client.Object
		expectedSHAAnnotation bool
	}{
		{
			name:            "scenario 1: no matching pods",
			secretNamespace: "test",
			secretName:      "testsecret",
			existingObjects: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testsecret",
						Namespace: "test",
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genPodWithSecretVolume("pod1", "test", "somesecret"),
			},

			expectedPods: []string{"pod1"},
		},
		{
			name:            "scenario 2: two matching pods, delete those pods",
			secretNamespace: "test",
			secretName:      "testsecret",
			existingObjects: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testsecret",
						Namespace: "test",
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genPodWithSecretVolume("pod1", "test", "somesecret"),
				genPodWithSecretVolume("pod2", "test", "testsecret"),
				genPodWithSecretVolume("pod3", "test", "somesecret"),
				genPodWithSecretVolume("pod4", "test", "testsecret"),
			},
			expectedPods: []string{"pod1", "pod3"},
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
			target := RedeploySecretReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("controllers").WithName("RedeploySecretsController"),
				Scheme: scheme.Scheme,
			}

			_, err := target.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: test.secretName, Namespace: test.secretNamespace}})
			assert.NoError(t, err)

			secret := &corev1.Secret{}
			err = fakeClient.Get(ctx, client.ObjectKey{Name: test.secretName, Namespace: test.secretNamespace}, secret)
			assert.NoError(t, err)

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

			assert.Equal(t, test.expectedPods, existingPodNames)
		})
	}
}

func genPodWithSecretVolume(podName, podNamespace, annotationSecretName string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        podName,
			Namespace:   podNamespace,
			Labels:      map[string]string{redeployment.RedeployLabel: "true"},
			Annotations: map[string]string{},
		},
	}
	if annotationSecretName != "" {
		pod.Annotations["security.plural.sh/oauth-env-secret"] = annotationSecretName
	}
	return pod
}
