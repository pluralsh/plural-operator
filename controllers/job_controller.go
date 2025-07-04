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
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobReconciler manages additional features on jobs
type JobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	expiresAfter    = 5 * time.Hour * 24
	managedByPlural = "plural"
)

//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Job object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("Job", req.NamespacedName)

	ns := req.Namespace
	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: ns}, &namespace); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch namespace")
		return ctrl.Result{}, err
	}

	if val, ok := namespace.Labels[managedLabel]; !ok || val != managedByPlural {
		log.Info(fmt.Sprintf("Namespace %s not managed by plural", ns))
		return ctrl.Result{}, nil
	}

	// your logic here
	var job batchv1.Job
	if err := r.Get(ctx, req.NamespacedName, &job); err != nil {
		log.Error(err, "Failed to fetch Job resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if job.Annotations != nil {
		if _, ok := job.Annotations[ignoreAnnotation]; ok {
			log.Info("Ignoring job due to annotation")
			return ctrl.Result{}, nil
		}
	}

	completion := job.Status.CompletionTime
	meta := job.ObjectMeta

	if completion == nil {
		log.Info("Job not yet completed")
		return ctrl.Result{}, nil
	}

	for _, owner := range meta.OwnerReferences {
		if owner.Kind == "CronJob" {
			log.Info(fmt.Sprintf("Job managed by CronJob %s, ignore", owner.Name))
			return ctrl.Result{}, nil
		}
	}

	dt := completion.Time
	expiry := dt.Add(expiresAfter)

	if time.Now().Before(expiry) {
		log.Info("Requeueing the job until after expiry")
		return ctrl.Result{RequeueAfter: time.Until(expiry)}, nil
	}

	if err := r.Delete(ctx, &job); err != nil {
		log.Error(err, "failed to delete job")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Successfully expired job")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).
		Complete(r)
}
