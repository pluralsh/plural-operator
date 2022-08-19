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
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/go-logr/logr"
	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RedeploySecretReconciler reconciles a Secret object for specified applications
type RedeploySecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *RedeploySecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pod", req.NamespacedName)

	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		log.Error(err, "Failed to fetch secret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	redeployments, err := r.getRedeploymentApplications(ctx, secret.Namespace)
	if err != nil {
		log.Error(err, "Failed to get redeployments")
		return ctrl.Result{}, err
	}

	deployments, err := r.getDeploymentsForSecret(ctx, redeployments, secret)
	if err != nil {
		log.Error(err, "Failed to get deployment for secret")
		return ctrl.Result{}, err
	}

	for _, deployment := range deployments {
		if secret.Annotations == nil {
			secret.Annotations = map[string]string{}
		}
		existingSHA, ok := secret.Annotations[shaAnnotation]
		expectedSHA := getSHA(secret)
		// create sha annotation when doesn't exist
		if !ok {
			secret.Annotations[shaAnnotation] = expectedSHA
			return ctrl.Result{}, r.Update(ctx, secret)
		}

		// restart deployment
		if existingSHA != expectedSHA {
			if deployment.Annotations == nil {
				deployment.Annotations = map[string]string{}
			}
			deployment.Annotations[deploymentRestartAnnotation] = time.Now().String()
			return ctrl.Result{}, r.Update(ctx, &deployment)
		}
	}

	return ctrl.Result{}, nil
}

func getSHA(secret *corev1.Secret) string {
	sha := sha256.New()
	for _, value := range secret.Data {
		sha.Write(value)
	}
	return hex.EncodeToString(sha.Sum(nil))
}

func (r *RedeploySecretReconciler) getRedeploymentApplications(ctx context.Context, namespace string) ([]platformv1alpha1.Redeployment, error) {
	redeploymentApplications := &platformv1alpha1.RedeploymentList{}

	if err := r.List(ctx, redeploymentApplications, &client.ListOptions{Namespace: namespace}); err != nil {
		return nil, err
	}

	return redeploymentApplications.Items, nil
}

func (r *RedeploySecretReconciler) getDeploymentsForSecret(ctx context.Context, redeployments []platformv1alpha1.Redeployment, secret *corev1.Secret) ([]appsv1.Deployment, error) {

	result := []appsv1.Deployment{}
	deployments := &appsv1.DeploymentList{}

	if err := r.List(ctx, deployments, &client.ListOptions{Namespace: secret.Namespace}); err != nil {
		return nil, err
	}
	for _, deployment := range deployments.Items {
		for _, redeployment := range redeployments {
			if redeployment.Spec.Workflow == platformv1alpha1.Deployment {
				for _, volume := range deployment.Spec.Template.Spec.Volumes {
					if volume.Secret != nil && volume.Secret.SecretName == secret.Name {
						result = append(result, deployment)
					}
				}
			}
		}
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedeploySecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
