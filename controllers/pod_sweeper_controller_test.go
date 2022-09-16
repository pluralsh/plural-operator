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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileSweeper(t *testing.T) {
	tests := []struct {
		name            string
		podName         string
		podNamespace    string
		expectedError   string
		expectedResult  ctrl.Result
		shouldDeletePod bool
		existingObjects []ctrlruntimeclient.Object
	}{
		{
			name:            "scenario 1: check and delete pod",
			podNamespace:    "test",
			podName:         "test",
			shouldDeletePod: true,
			expectedResult:  ctrl.Result{},
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: corev1.PodSpec{},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
						Conditions: []corev1.PodCondition{
							{Type: corev1.PodReady, Status: corev1.ConditionFalse, LastTransitionTime: metav1.NewTime(Now().Add(-10 * time.Minute))},
						},
					},
				},
			},
		},
		{
			name:            "scenario 2: pod is pending, check later",
			podNamespace:    "test",
			podName:         "test",
			shouldDeletePod: true,
			expectedResult:  ctrl.Result{RequeueAfter: 6 * time.Minute},
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: corev1.PodSpec{},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
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
			target := PodSweeperReconciler{
				Client:      fakeClient,
				Log:         ctrl.Log.WithName("controllers").WithName("PodSweeperController"),
				Scheme:      scheme.Scheme,
				DeleteAfter: 5 * time.Minute,
			}

			result, err := target.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: test.podName, Namespace: test.podNamespace}})
			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				err := fakeClient.Get(ctx, ctrlruntimeclient.ObjectKey{Namespace: test.podNamespace, Name: test.name}, &corev1.Pod{})
				if test.shouldDeletePod {
					err = ctrlruntimeclient.IgnoreNotFound(err)
				}
				assert.NoError(t, err)
				assert.Equal(t, result, test.expectedResult)
			}

		})
	}
}
