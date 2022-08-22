package controllers

import (
	"context"
	"testing"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural-operator/services/redeployment"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
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

func TestReconcileSecret(t *testing.T) {
	tests := []struct {
		name                   string
		secretName             string
		secretNamespace        string
		deploymentsForRestart  []string
		statefulSetsForRestart []string
		daemonSetsForRestart   []string
		existingObjects        []ctrlruntimeclient.Object
		expectedSHAAnnotation  bool
	}{
		{
			name:            "scenario 1: no redeployments, don't add SHA annotation",
			secretNamespace: "test",
			secretName:      "testsecret",
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testsecret",
						Namespace: "test",
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genDeploymentWithSecretVolume("dep1", "test", "testsecret"),
			},
		},
		{
			name:                  "scenario 2: add SHA annotation to the secret, when doesn't exist",
			secretNamespace:       "test",
			secretName:            "testsecret",
			expectedSHAAnnotation: true,
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testsecret",
						Namespace: "test",
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genRedeployment("dep1", "test", v1alpha1.Deployment),
				genRedeployment("dep3", "test", v1alpha1.Deployment),
				genDeploymentWithSecretVolume("dep1", "test", "testsecret"),
			},
		},
		{
			name:                  "scenario 3: restart only deployments after secret changes",
			secretNamespace:       "test",
			secretName:            "testsecret",
			expectedSHAAnnotation: true,
			deploymentsForRestart: []string{"dep1", "dep3"},
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "testsecret",
						Namespace:   "test",
						Annotations: map[string]string{redeployment.ShaAnnotation: "xyz"},
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genRedeployment("dep1", "test", v1alpha1.Deployment),
				genRedeployment("dep3", "test", v1alpha1.Deployment),
				genDeploymentWithSecretVolume("dep1", "test", "testsecret"),
				genDeploymentWithSecretVolume("dep2", "test", "sometest"),
				genDeploymentWithSecretVolume("dep3", "test", "testsecret"),
				genStatefulSetWithSecretVolume("state2", "test", "sometest"),
				genDeamonSetWithSecretVolume("daemon3", "test", "sometest"),
			},
		},
		{
			name:                   "scenario 4: restart deployments, daemonSets, and statefulSets after secret changes",
			secretNamespace:        "test",
			secretName:             "testsecret",
			expectedSHAAnnotation:  true,
			deploymentsForRestart:  []string{"dep1", "dep3"},
			daemonSetsForRestart:   []string{"daemon1", "daemon2"},
			statefulSetsForRestart: []string{"state1"},
			existingObjects: []ctrlruntimeclient.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "testsecret",
						Namespace:   "test",
						Annotations: map[string]string{redeployment.ShaAnnotation: "xyz"},
					},
					Data: map[string][]byte{"z": {1, 2, 3}, "a": {4, 5, 6}},
				},
				genRedeployment("dep1", "test", v1alpha1.Deployment),
				genRedeployment("dep3", "test", v1alpha1.Deployment),
				genDeploymentWithSecretVolume("dep1", "test", "testsecret"),
				genDeploymentWithSecretVolume("dep2", "test", "sometest"),
				genDeploymentWithSecretVolume("dep3", "test", "testsecret"),
				genStatefulSetWithSecretVolume("state1", "test", "testsecret"),
				genStatefulSetWithSecretVolume("state2", "test", "sometest"),
				genDeamonSetWithSecretVolume("daemon1", "test", "testsecret"),
				genDeamonSetWithSecretVolume("daemon2", "test", "testsecret"),
				genDeamonSetWithSecretVolume("daemon3", "test", "sometest"),
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

			if test.expectedSHAAnnotation {
				_, shaAnnotation := secret.Annotations[redeployment.ShaAnnotation]
				assert.True(t, shaAnnotation, "expected SHA annotation")
			}

			for _, deployment := range test.deploymentsForRestart {
				d := &appsv1.Deployment{}
				err := fakeClient.Get(ctx, client.ObjectKey{Name: deployment, Namespace: test.secretNamespace}, d)
				assert.NoError(t, err)
				if d.Spec.Template.ObjectMeta.Annotations == nil {
					t.Fatalf("expected annotations for deployment %s", deployment)
				}
				_, restartAnnotation := d.Spec.Template.ObjectMeta.Annotations[redeployment.RestartAnnotation]
				assert.True(t, restartAnnotation, "expected restart annotation")
			}
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

func genDeamonSetWithSecretVolume(name, namespace, secretName string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{
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

func genStatefulSetWithSecretVolume(name, namespace, secretName string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
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

func genRedeployment(name, namespace string, workflow v1alpha1.WorkflowType) *v1alpha1.Redeployment {
	return &v1alpha1.Redeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.RedeploymentSpec{
			Name:      name,
			Namespace: namespace,
			Workflow:  workflow,
		},
	}
}
