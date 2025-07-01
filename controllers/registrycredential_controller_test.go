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
	"encoding/base64"
	"testing"

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
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

func TestReconcileRegistryCredential(t *testing.T) {
	tests := []struct {
		name            string
		registryCred    *v1alpha1.RegistryCredential
		expectedAuths   string
		expectedError   string
		existingObjects []client.Object
	}{
		{
			name:         "scenario 1: create new secret with credentials",
			registryCred: genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			existingObjects: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{"password": []byte("cGFzc3dvcmQ=")},
				},
				genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			},
			expectedAuths: "{\"auths\":{\"dkr.plural.sh\":{\"auth\":\"dGVzdDpjR0Z6YzNkdmNtUT0=\",\"password\":\"cGFzc3dvcmQ=\",\"username\":\"test\",\"email\":\"test@plural.sh\"}}}",
		},
		{
			name:         "scenario 2: update registry credentials",
			registryCred: genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			existingObjects: []client.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret",
						Namespace: "test",
					},
					Data: map[string][]byte{"password": []byte("bmV3")},
				},
				// existing secret
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "registry-secret-cred",
						Namespace: "test",
					},
					Data: map[string][]byte{dockerconfigjson: []byte("eyJhdXRocyI6eyJka3IucGx1cmFsLnNoIjp7ImF1dGgiOiJkR1Z6ZERwalIwWjZZek5rZG1OdFVUMD0iLCJwYXNzd29yZCI6ImNHRnpjM2R2Y21RPSIsInVzZXJuYW1lIjoidGVzdCIsImVtYWlsIjoidGVzdEBwbHVyYWwuc2gifX19")},
				},
				genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			},

			expectedAuths: "{\"auths\":{\"dkr.plural.sh\":{\"auth\":\"dGVzdDpibVYz\",\"password\":\"bmV3\",\"username\":\"test\",\"email\":\"test@plural.sh\"}}}",
		},
		{
			name:         "scenario 3: password secret doesn't exist",
			registryCred: genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			existingObjects: []client.Object{
				genRegistryCred("cred", "test", "secret", "password", "test", "test@plural.sh", "dkr.plural.sh"),
			},
			expectedError: "secrets \"secret\" not found",
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
			target := RegistryCredentialsReconciler{
				Client: fakeClient,
				Log:    ctrl.Log.WithName("controllers").WithName("RegistryCredentialsReconciler"),
				Scheme: scheme.Scheme,
			}

			_, err := target.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: test.registryCred.Name, Namespace: test.registryCred.Namespace}})
			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)

				credSecret := &corev1.Secret{}
				err = fakeClient.Get(ctx, client.ObjectKey{Namespace: test.registryCred.Namespace, Name: test.registryCred.GetPasswordSecretName()}, credSecret)
				assert.NoError(t, err)

				auths := credSecret.Data[dockerconfigjson]
				currentAuths, _ := base64.StdEncoding.DecodeString(string(auths))
				assert.Equal(t, test.expectedAuths, string(currentAuths))
			}
		})
	}
}

func genRegistryCred(name, namespace, secretName, secretKey, username, email, server string) *v1alpha1.RegistryCredential {
	return &v1alpha1.RegistryCredential{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.RegistryCredentialSpec{
			Username: username,
			Email:    email,
			Server:   server,
			PasswordSecretRef: v1alpha1.PasswordSecretRef{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}
