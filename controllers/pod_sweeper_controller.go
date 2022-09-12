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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Now stubbed out to allow testing.
var Now = time.Now

const defaultSweepAfter = 2 * time.Hour

// PodSweeperReconciler reconciles a Pod object
type PodSweeperReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	DeleteAfter time.Duration
}

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcilee
func (r *PodSweeperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("pod", req.NamespacedName)

	if r.DeleteAfter == 0 {
		r.DeleteAfter = defaultSweepAfter
	}

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get a standalone pod
	// Skip all k8s workflow pods. The owner reference for the workflow pod is always set with the specific kind: ReplicaSet, DaemonSet
	if pod.OwnerReferences != nil {
		return ctrl.Result{}, nil
	}

	// Skip running pods, and check them later
	if pod.Status.Phase == corev1.PodRunning {
		return ctrl.Result{
			RequeueAfter: r.DeleteAfter,
		}, nil
	}

	// If the condition is not set yet then set lastNotReadyTimestamp to Now() to enforce the check for later
	now := Now()
	lastNotReadyTimestamp := metav1.Time{now}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			lastNotReadyTimestamp = condition.LastTransitionTime
		}
	}

	notReadyDuration := now.Sub(lastNotReadyTimestamp.Time)
	checkDuration := r.DeleteAfter
	if notReadyDuration >= checkDuration {
		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "Failed to delete pod")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}

	repeatAfter := checkDuration.Truncate(notReadyDuration)
	return ctrl.Result{RequeueAfter: repeatAfter + time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodSweeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.GetControllerOptions()
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
