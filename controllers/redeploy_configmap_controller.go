package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/pluralsh/plural-operator/services/redeployment"
)

type ConfigMapRedeployReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (c *ConfigMapRedeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := c.Log.WithValues("reconcile", req.NamespacedName)

	configMap := &corev1.ConfigMap{}
	err := c.Client.Get(ctx, req.NamespacedName, configMap)
	if errors.IsNotFound(err) {
		log.Error(nil, "could not find configmap")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ConfigMap: %+v", err)
	}

	svc := redeployment.NewFactory().Create(redeployment.ResourceConfigMap, c.Client, configMap)
	controlled, err := svc.IsControlled()
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not check if config map is controlled by any application: %+v", err)
	}

	if !controlled {
		return reconcile.Result{}, nil
	}

	if !svc.HasAnnotation() {
		log.Info("updating config map with new sha")
		return reconcile.Result{}, svc.UpdateAnnotation()
	}

	if svc.ShouldRestart() {
		err = svc.UpdateAnnotation()
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("could not update annotation: %+v", err)
		}

		log.Info("starting a rollout restart")
		return reconcile.Result{}, svc.RolloutRestart()
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *ConfigMapRedeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(c)
}