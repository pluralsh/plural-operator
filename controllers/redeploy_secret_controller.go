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
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	"github.com/pluralsh/plural-operator/services/redeployment"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

	svc, err := redeployment.NewSecretService(r.Client, secret)
	if err != nil {
		log.Error(nil, "could not create Secret service")
		return reconcile.Result{}, err
	}

	if !svc.IsControlled() {
		return reconcile.Result{}, nil
	}

	log.Info("deleting Pods")
	err = svc.DeletePods()
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not delete pods: %w", err)
	}

	return ctrl.Result{}, nil
}

func getSecretSHA(s *corev1.Secret) string {
	sha := sha256.New()
	dataKeys := make([]string, 0)

	for key := range s.Data {
		dataKeys = append(dataKeys, key)
	}

	sort.Strings(dataKeys)

	for _, key := range dataKeys {
		sha.Write(s.Data[key])
	}

	return hex.EncodeToString(sha.Sum(nil))
}

func updateSecretEventsOnly() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldSecret, ok := e.ObjectOld.(*corev1.Secret)
			if !ok {
				return false
			}
			newSecret, ok := e.ObjectNew.(*corev1.Secret)
			if !ok {
				return false
			}
			oldSHA := getSecretSHA(oldSecret)
			newSHA := getSecretSHA(newSecret)
			return oldSHA != newSHA
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedeploySecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(updateSecretEventsOnly()).
		Complete(r)
}
