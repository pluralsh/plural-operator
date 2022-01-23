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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodReconciler reconciles a Namespace object
type PodReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	expiryAnnotation = "platform.plural.sh/expire-after"
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
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pod", req.NamespacedName)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		log.Error(err, "Failed to fetch pod")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if pod.Annotations == nil {
		return ctrl.Result{}, nil
	}

	expiry, ok := pod.Annotations[expiryAnnotation]
	if !ok {
		return ctrl.Result{}, nil
	}

	log.Info("Found expiry annotation", "expiry", expiry)
	dur, err := time.ParseDuration(expiry)
	if err != nil {
		log.Error(err, "Failed to parse expiry duration")
		return ctrl.Result{}, nil
	}

	created := pod.CreationTimestamp
	expiresAt := created.Add(dur)
	if expiresAt.Before(time.Now()) {
		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "Failed to delete pod")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: time.Until(expiresAt)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
