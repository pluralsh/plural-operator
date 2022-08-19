package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ctrl "sigs.k8s.io/controller-runtime"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileForDeployments(t *testing.T) {
	tests := []struct {
		name            string
		secretName      string
		secretNamespace string
		existingObjects []ctrlruntimeclient.Object
	}{
		{
			name: "scenario 1: add sha annotation for the secret",

			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testsecret",
						Namespace: "test",
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genDeploymentWithSecretVolume("dep1", "test", "testsecret"),
				genDeploymentWithSecretVolume("dep2", "test", "sometest"),
				genDeploymentWithSecretVolume("dep3", "test", "testsecret"),
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

		})
	}
}

func genDeploymentWithSecretVolume(deploymentName, deploymentNamespace, secretName string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: deploymentNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secretName,
								},
							},
						},
					},
				},
			},
		},
	}
}
