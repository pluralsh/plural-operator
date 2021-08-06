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
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

// SecretSyncReconciler reconciles a SecretSync object
type SecretSyncReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	secretOwnerKey    = ".metadata.controller"
	apiGVStr          = platformv1alpha1.GroupVersion.String()
	ownedLabel        = "platform.plural.sh/owned"
	ownerAnnotation   = "platform.plural.sh/owner"
	allowedAnnotation = "platform.plural.sh/syncable"
)

//+kubebuilder:rbac:groups=platform.plural.sh,resources=secretsyncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.plural.sh,resources=secretsyncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.plural.sh,resources=secretsyncs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *SecretSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secretsync", req.NamespacedName)

	// your logic here
	var sync platformv1alpha1.SecretSync
	if err := r.Get(ctx, req.NamespacedName, &sync); err != nil {
		log.Error(err, "Failed to fetch secret sync resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var secret corev1.Secret
	namespacedName := types.NamespacedName{
		Namespace: sync.Spec.Namespace,
		Name:      sync.Spec.Name,
	}
	if err := r.Get(ctx, namespacedName, &secret); err != nil {
		log.Error(err, "Could not fetch secret to sync")
		return ctrl.Result{}, err
	}

	meta := secret.ObjectMeta
	// if _, ok := secret.Annotations[allowedAnnotation]; !ok {
	// 	log.Info("Secret is not labeled as syncable: ", meta.Namespace, "/", meta.Name)
	// 	return ctrl.Result{}, nil
	// }

	oldNs := meta.Namespace

	secret.ObjectMeta.Namespace = sync.ObjectMeta.Namespace
	if err := r.Patch(ctx, &secret, client.Apply, client.ForceOwnership, client.FieldOwner("plural-operator")); err != nil {
		log.Error(err, "failed to sync object to namespace", sync.ObjectMeta.Namespace)
		return ctrl.Result{}, err
	}

	secret.Labels[ownedLabel] = "true"
	secret.Annotations[ownerAnnotation] = fmt.Sprintf("%s/%s", req.NamespacedName.Namespace, req.NamespacedName.Name)
	secret.ObjectMeta.Namespace = oldNs
	if err := r.Update(ctx, &secret); err != nil {
		log.Error(err, "Failed to add ownership labels to secret: ", secret.ObjectMeta.Namespace, "/", secret.ObjectMeta.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.SecretSync{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			secret := obj.(*corev1.Secret)
			if val, ok := secret.Labels[ownedLabel]; !ok || val != "true" {
				return []reconcile.Request{}
			}

			owner := strings.Split(secret.Annotations[ownerAnnotation], "/")
			if len(owner) != 2 {
				return []reconcile.Request{}
			}

			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Namespace: owner[0],
					Name:      owner[1],
				}},
			}
		})).
		Complete(r)
}
