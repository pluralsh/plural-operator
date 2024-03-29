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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
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
	ownedAnnotation   = "platform.plural.sh/owned"
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
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch secret sync resource")
		return ctrl.Result{}, err
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

	// if _, ok := secret.Annotations[allowedAnnotation]; !ok {
	// 	log.Info("Secret is not labeled as syncable: ", meta.Namespace, "/", meta.Name)
	// 	return ctrl.Result{}, nil
	// }

	meta := secret.ObjectMeta
	log.Info(fmt.Sprintf("Attempting to apply secret %s/%s", req.NamespacedName.Namespace, meta.Name))
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}

	if !r.isSyncNecessary(ctx, &secret, &sync) {
		log.Info("Secret already exists in target namespace and is correctly formed")
		return ctrl.Result{}, nil
	}

	annotations := secret.Annotations
	delete(annotations, ownerAnnotation)
	delete(annotations, ownedAnnotation)
	newSecret := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":        meta.Name,
				"namespace":   req.NamespacedName.Namespace,
				"labels":      secret.Labels,
				"annotations": annotations,
			},
			"data": secret.Data,
		},
	}
	if err := r.Patch(ctx, newSecret, client.Apply, client.ForceOwnership, client.FieldOwner("plural-operator")); err != nil {
		log.Error(err, fmt.Sprintf("failed to sync object to namespace %s", sync.ObjectMeta.Namespace))
		return ctrl.Result{}, err
	}

	secret.Annotations[ownedAnnotation] = "true"
	secret.Annotations[ownerAnnotation] = fmt.Sprintf("%s/%s", req.NamespacedName.Namespace, req.NamespacedName.Name)
	if err := r.Update(ctx, &secret); err != nil {
		log.Error(err, fmt.Sprintf("Failed to add ownership labels to secret: %s/%s", secret.ObjectMeta.Namespace, secret.ObjectMeta.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SecretSyncReconciler) isSyncNecessary(ctx context.Context, secret *corev1.Secret, sync *platformv1alpha1.SecretSync) bool {
	var target corev1.Secret
	namespacedName := types.NamespacedName{
		Namespace: sync.Namespace,
		Name:      secret.Name,
	}
	if err := r.Get(ctx, namespacedName, &target); err != nil {
		r.Log.Info("No target sync found, sync is required")
		return true
	}

	return !reflect.DeepEqual(target.Data, secret.Data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.SecretSync{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			secret := obj.(*corev1.Secret)
			if val, ok := secret.Annotations[ownedAnnotation]; !ok || val != "true" {
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
