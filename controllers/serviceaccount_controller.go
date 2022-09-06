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

package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const pluralCreds = "plural-creds"

// ServiceAccountReconciler reconciles a SecretSync object
type ServiceAccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("serviceaccount", req.NamespacedName)

	ns := req.NamespacedName.Namespace
	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: ns}, &namespace); err != nil {
		log.Error(err, "Failed to fetch namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if val, ok := namespace.Labels[managedLabel]; !ok || val != "plural" {
		log.Info(fmt.Sprintf("Namespace %s not managed by plural", ns))
		return ctrl.Result{}, nil
	}

	// your logic here
	var serviceaccount corev1.ServiceAccount
	if err := r.Get(ctx, req.NamespacedName, &serviceaccount); err != nil {
		log.Error(err, "Failed to fetch serviceaccount resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get other credential secrets
	labelSelector, err := pluralCredLabelSelector()
	if err != nil {
		log.Error(err, "Failed to fetch label selector")
		return ctrl.Result{}, err
	}
	credSecrets := &corev1.SecretList{}
	if err := r.List(ctx, credSecrets, &client.ListOptions{Namespace: serviceaccount.Namespace, LabelSelector: labelSelector}); err != nil {
		log.Error(err, "Failed to fetch credential secrets")
		return ctrl.Result{}, err
	}

	existingPullSecretsSet := sets.NewString()
	expectedPullSecretsSet := sets.NewString(pluralCreds)

	for _, secret := range serviceaccount.ImagePullSecrets {
		existingPullSecretsSet.Insert(secret.Name)
	}
	for _, secret := range credSecrets.Items {
		expectedPullSecretsSet.Insert(secret.Name)
	}
	if existingPullSecretsSet.HasAll(expectedPullSecretsSet.List()...) {
		log.Info("Service account already has creds attached")
		return ctrl.Result{}, nil
	}

	serviceaccount.ImagePullSecrets = []corev1.LocalObjectReference{}

	for _, secret := range expectedPullSecretsSet.List() {
		serviceaccount.ImagePullSecrets = append(serviceaccount.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secret,
		})
	}

	if err := r.Update(ctx, &serviceaccount); err != nil {
		meta := serviceaccount.ObjectMeta
		log.Error(err, fmt.Sprintf("Failed to add pullsecrets to sa: %s/%s", meta.Namespace, meta.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func pluralCredLabelSelector() (labels.Selector, error) {
	req, err := labels.NewRequirement(secretTypeLabel, selection.Equals, []string{pluralCreds})
	if err != nil {
		return nil, fmt.Errorf("failed to build label selector: %w", err)
	}
	return labels.Parse(req.String())
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}
