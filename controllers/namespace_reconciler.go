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

	"github.com/go-logr/logr"
	"github.com/pluralsh/plural-operator/resources"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
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
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("namespace", req.NamespacedName)

	ns := req.NamespacedName.Name
	var namespace corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: ns}, &namespace); err != nil {
		log.Error(err, "Failed to fetch namespace")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if val, ok := namespace.Labels[managedLabel]; !ok || val != "plural" {
		log.Info(fmt.Sprintf("Namespace %s not managed by plural", ns))
		return ctrl.Result{}, nil
	}

	pluralAlerts := resources.AlertManagerConfig(
		"plural",
		"http://plural-operator.bootstrap:8080/webhook",
		map[string]string{},
	)

	consoleAlerts := resources.AlertManagerConfig(
		"console",
		"http://console.console:4000/alertmanager",
		map[string]string{},
	)

	for _, obj := range []*v1alpha1.AlertmanagerConfig{pluralAlerts, consoleAlerts} {
		obj.SetNamespace(namespace.Name)

		log.Info("Applying alertmanager config", "name", obj.Name)
		spec := obj.Spec
		if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, func() error {
			obj.Spec = spec
			return nil
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
