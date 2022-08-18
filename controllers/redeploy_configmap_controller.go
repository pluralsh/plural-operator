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

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type ConfigMapRedeployReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

func (c *ConfigMapRedeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := c.Log.WithValues("redeploy", req.NamespacedName)

	redeploymentList := &v1alpha1.RedeploymentList{}
	err := c.Client.List(ctx, redeploymentList, &client.ListOptions{})
	if errors.IsNotFound(err) {
		log.Error(nil, "could not find RedeploymentList")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch RedeploymentList: %+v", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (c *ConfigMapRedeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(c)
}
