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
)

type ConfigMapRedeployReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	Log                logr.Logger
	ConfigMapName      string
	ConfigMapNamespace string
}

func (c *ConfigMapRedeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := c.Log.WithValues("reconcile", req.NamespacedName)

	configMap := &corev1.ConfigMap{}
	if err := c.Get(ctx, req.NamespacedName, configMap); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch config map")
		return ctrl.Result{}, err
	}

	redeployLabelSelector, err := redeployment.RedeployLabelSelector()
	if err != nil {
		return ctrl.Result{}, err
	}

	pods := &corev1.PodList{}
	if err := c.List(ctx, pods, &client.ListOptions{LabelSelector: redeployLabelSelector}); err != nil {
		return ctrl.Result{}, err
	}
	for _, pod := range pods.Items {
		if err := c.Delete(ctx, &pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func updateConfigMapEventsOnly() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConfigMap, ok := e.ObjectOld.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			newConfigMap, ok := e.ObjectNew.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			oldSHA := getConfigMapSHA(oldConfigMap)
			newSHA := getConfigMapSHA(newConfigMap)
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

func getConfigMapSHA(c *corev1.ConfigMap) string {
	sha := sha256.New()
	dataKeys := make([]string, 0)
	binaryDataKeys := make([]string, 0)

	for key := range c.Data {
		dataKeys = append(dataKeys, key)
	}

	for key := range c.BinaryData {
		binaryDataKeys = append(binaryDataKeys, key)
	}

	sort.Strings(dataKeys)
	sort.Strings(binaryDataKeys)

	for _, key := range dataKeys {
		sha.Write([]byte(c.Data[key]))
	}

	for _, key := range binaryDataKeys {
		sha.Write(c.BinaryData[key])
	}

	return hex.EncodeToString(sha.Sum(nil))
}

// SetupWithManager sets up the controller with the Manager.
func (c *ConfigMapRedeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	byNameAndNamespace := predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetName() == c.ConfigMapName && object.GetNamespace() == c.ConfigMapNamespace
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(byNameAndNamespace).
		WithEventFilter(updateConfigMapEventsOnly()).
		Complete(c)
}
