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

	"github.com/go-logr/logr"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural-operator/resources"
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
//+kubebuilder:rbac:groups="storage.k8s.io",resources=storageclasses,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the DefaultStorageClass object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *DefaultStorageClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("defaultstorageclass", req.NamespacedName)

	defaultstorageInstance := &platformv1alpha1.DefaultStorageClass{}

	if err := r.Get(ctx, req.NamespacedName, defaultstorageInstance); err != nil {
		log.Error(err, "could not fetch default storage class resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var foundStorageClass storagev1.StorageClass

	// Update storageClass
	// Update the storage class so that it becomes the default
	storageClass := r.generateStorageClass(defaultstorageInstance)
	// Don't continue if the storage class does not exist. Without this check resources.StorageClass would create the storage class
	if err := r.Get(ctx, types.NamespacedName{Name: defaultstorageInstance.Spec.Name}, &foundStorageClass); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("storage class does not exist", "class", defaultstorageInstance.Spec.Name)
			return ctrl.Result{}, nil
		}
	} else if err := resources.StorageClass(ctx, r.Client, storageClass, log); err != nil {
		log.Error(err, "Error reconciling default StorageClass", "class", storageClass.Name)
		return ctrl.Result{}, err
	}

	var classes storagev1.StorageClassList

	// Get the list of storage classes from the cluster
	if err := r.List(ctx, &classes); err != nil {
		log.Error(err, "could not list storage classes")
		return ctrl.Result{}, err
	}

	for _, class := range classes.Items {
		// fmt.Printf("%+v", class)
		if _, ok := class.Annotations[defaultStorageAnnotation]; ok && class.Name != defaultstorageInstance.Spec.Name {
			log.Info("setting storage class to non-default", "class", class.Name)
			delete(class.Annotations, defaultStorageAnnotation)
			if err := r.Update(ctx, &class); err != nil {
				log.Error(err, "failed to update previous default", "storageclass", class.Name)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DefaultStorageClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.DefaultStorageClass{}).
		Complete(r)
}

// create a dummy storage class object
func (r *DefaultStorageClassReconciler) generateStorageClass(defaultstorageInstance *platformv1alpha1.DefaultStorageClass) *storagev1.StorageClass {
	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultstorageInstance.Spec.Name,
			Annotations: map[string]string{defaultStorageAnnotation: "true"},
		},
	}
	return storageClass
}
