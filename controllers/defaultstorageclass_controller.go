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
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

const (
	defaultStorageAnnotation = "storageclass.kubernetes.io/is-default-class"
)

// DefaultStorageClassReconciler reconciles a DefaultStorageClass object
type DefaultStorageClassReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=plaform.plural.sh.plural.sh,resources=defaultstorageclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=plaform.plural.sh.plural.sh,resources=defaultstorageclasses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=plaform.plural.sh.plural.sh,resources=defaultstorageclasses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DefaultStorageClass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DefaultStorageClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("defaultstorageclass", req.NamespacedName)

	var defaultstorage platformv1alpha1.DefaultStorageClass

	if req.NamespacedName.Name != "default" {
		err := fmt.Errorf("Cannot reconcile default storage class resource with name %s", req.NamespacedName.Name)
		log.Error(err, "forcibly ignoring invalid name", "name", req.NamespacedName.Name)
		return ctrl.Result{}, err
	}

	if err := r.Get(ctx, req.NamespacedName, &defaultstorage); err != nil {
		log.Error(err, "could not fetch default storage class resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	classes := &storagev1.StorageClassList{}
	var storageClass *storagev1.StorageClass
	var currDefault *storagev1.StorageClass
	if err := r.List(ctx, classes); err != nil {
		log.Error(err, "could not list storage classes")
		return ctrl.Result{}, err
	}

	for _, class := range classes.Items {
		fmt.Printf("%+v", class)
		if _, ok := class.Annotations[defaultStorageAnnotation]; ok {
			currDefault = &class
		}

		if class.Name == defaultstorage.Spec.Name {
			storageClass = &class
		}
	}

	if storageClass == nil {
		err := fmt.Errorf("Could not find storage class matching %s", defaultstorage.Spec.Name)
		log.Error(err, "failed to find storage class")
		return ctrl.Result{}, err
	}

	if currDefault != nil && currDefault.Name != defaultstorage.Spec.Name {
		delete(currDefault.Annotations, defaultStorageAnnotation)
		if err := r.Update(ctx, currDefault); err != nil {
			log.Error(err, "failed to update previous default", "storageclass", currDefault.Name)
			return ctrl.Result{}, err
		}
	}

	storageClass.Annotations[defaultStorageAnnotation] = "true"
	if err := r.Update(ctx, storageClass); err != nil {
		log.Error(err, "failed to modify default storage class", "storageclass", storageClass.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DefaultStorageClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.DefaultStorageClass{}).
		Complete(r)
}
