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
	"time"

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
)

// SecretReconciler reconciles a Namespace object
type SecretReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	secretSyncLabel      = "platform.plural.sh/sync"
	namespaceBucketLabel = "platform.plural.sh/sync-target"
	allBuckets           = "all"
)

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
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secret", req.NamespacedName)

	var secret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch secret")
		return ctrl.Result{}, err
	}

	if secret.Labels == nil {
		return ctrl.Result{}, nil
	}

	bucket, ok := secret.Labels[secretSyncLabel]
	if !ok {
		return ctrl.Result{}, nil
	}

	listOpts := &client.ListOptions{}
	if bucket == allBuckets {
		match := client.MatchingLabels{}
		match[managedLabel] = managedByPlural
		match.ApplyToList(listOpts)
	} else {
		hasLabels := client.HasLabels{namespaceBucketLabel}
		hasLabels.ApplyToList(listOpts)
	}

	namespaces := &corev1.NamespaceList{}
	if err := r.List(ctx, namespaces, listOpts); err != nil {
		log.Error(err, "could not list namespaces")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Attempting to sync secret across namespaces", "secret", secret.Name)
	for _, ns := range namespaces.Items {
		if ns.Name == secret.Namespace {
			continue
		}

		val := ns.Labels[namespaceBucketLabel]
		buckets := strings.Split(val, ",")
		for _, b := range buckets {
			if b == bucket || bucket == allBuckets {
				log.Info("Syncing to namespace", "namespace", ns.Name)
				if err := r.syncSecret(ctx, &secret, &ns); err != nil {
					return ctrl.Result{}, err
				}
				break
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *SecretReconciler) isSyncNecessary(ctx context.Context, secret *corev1.Secret, ns *corev1.Namespace) bool {
	var target corev1.Secret
	namespacedName := types.NamespacedName{
		Namespace: ns.Name,
		Name:      secret.Name,
	}
	if err := r.Get(ctx, namespacedName, &target); err != nil {
		r.Log.Info("No target sync found, sync is required")
		return true
	}

	return !reflect.DeepEqual(target.Data, secret.Data)
}

func (r *SecretReconciler) syncSecret(ctx context.Context, secret *corev1.Secret, ns *corev1.Namespace) error {
	log := r.Log.WithValues("namespace", secret.Namespace)
	labels := secret.Labels
	delete(labels, secretSyncLabel)
	newSecret := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":        secret.Name,
				"namespace":   ns.Name,
				"labels":      labels,
				"annotations": secret.Annotations,
			},
			"type": secret.Type,
			"data": secret.Data,
		},
	}

	if !r.isSyncNecessary(ctx, secret, ns) {
		log.Info("Secret already synced to namespace")
		return nil
	}

	if err := r.Patch(ctx, newSecret, client.Apply, client.ForceOwnership, client.FieldOwner("plural-operator")); err != nil {
		log.Error(err, fmt.Sprintf("failed to sync secret %s to namespace %s", secret.Name, ns.Namespace))
		return err
	}

	log.Info(fmt.Sprintf("Synced secret %s to namespace %s", secret.Name, ns.Namespace))
	return nil
}

func (r *SecretReconciler) listEffectedSecretsForNamespace(object client.Object) []reconcile.Request {
	ns := object.(*corev1.Namespace)
	log := r.Log.WithValues("namespace", ns.Name)
	log.Info("Found namespace update")
	results := []reconcile.Request{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if ns.Labels == nil {
		log.Info("namespace has no labels")
		return results
	}

	if res, ok := ns.Labels[managedLabel]; !ok || res != managedByPlural {
		log.Info("namespace not managed by plural")
		return results
	}

	val := ns.Labels[namespaceBucketLabel]
	buckets := strings.Split(val, ",")
	secrets := &corev1.SecretList{}
	if err := r.List(ctx, secrets, client.HasLabels{secretSyncLabel}); err != nil {
		return results
	}

	log.Info(fmt.Sprintf("Found %d secrets to potentially sync", len(secrets.Items)))

	for _, secret := range secrets.Items {
		bucket := secret.Labels[secretSyncLabel]
		if bucket == allBuckets {
			log.Info("Adding secret to update queue", "name", secret.Name, "namespace", secret.Namespace)
			results = append(results, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&secret),
			})
			continue
		}

		for _, b := range buckets {
			if b == bucket {
				log.Info("Adding secret to update queue", "name", secret.Name, "namespace", secret.Namespace)
				results = append(results, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&secret),
				})
				break
			}
		}
	}

	return results
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Watches(
			&source.Kind{Type: &corev1.Namespace{}},
			handler.EnqueueRequestsFromMapFunc(r.listEffectedSecretsForNamespace),
		).
		Complete(r)
}
