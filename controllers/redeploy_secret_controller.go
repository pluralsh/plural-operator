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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pluralsh/plural-operator/services/redeployment"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	log := r.Log.WithValues("secret", req.NamespacedName)

	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch secret")
		return ctrl.Result{}, err
	}

	svc, err := redeployment.NewService(redeployment.ResourceSecret, r.Client, secret)
	if err != nil {
		log.Error(nil, "could not create Secret service")
		return reconcile.Result{}, err
	}

	if !svc.IsControlled() {
		return reconcile.Result{}, nil
	}

	if !svc.HasAnnotation() {
		log.Info("update secret annotation")
		return reconcile.Result{}, svc.UpdateAnnotation()
	}

	if svc.ShouldDeletePods() {
		log.Info("deleting Pods")
		err = svc.DeletePods()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("could not delete pods: %w", err)
		}

		err = svc.UpdateAnnotation()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("could not update annotation: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedeploySecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, redeployment.RequestSecretFromPod(mgr.GetClient())).
		Complete(r)
}
