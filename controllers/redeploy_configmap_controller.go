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

type ConfigMapRedeployReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (c *ConfigMapRedeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := c.Log.WithValues("reconcile", req.NamespacedName)

	configMap := &corev1.ConfigMap{}
	if err := c.Client.Get(ctx, req.NamespacedName, configMap); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch config map")
		return ctrl.Result{}, err
	}

	svc, err := redeployment.NewService(redeployment.ResourceConfigMap, c.Client, configMap)
	if err != nil {
		log.Error(nil, "could not create ConfigMap service")
		return reconcile.Result{}, err
	}

	if !svc.IsControlled() {
		return reconcile.Result{}, nil
	}

	if !svc.HasAnnotation() {
		log.Info("update config map annotation")
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
func (c *ConfigMapRedeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, redeployment.RequestConfigMapFromPod(mgr.GetClient())).
		Complete(c)
}
