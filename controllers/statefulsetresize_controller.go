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
	"github.com/pluralsh/plural-operator/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

// StatefulSetResizeReconciler reconciles a StatefulSetResize object
type StatefulSetResizeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *StatefulSetResizeReconciler) cleanup(ctx context.Context, resize *platformv1alpha1.StatefulSetResize) (ctrl.Result, error) {
	log := r.Log.WithValues("statefulsetresize", resize.Name)
	if err := r.Delete(ctx, resize); err != nil {
		log.Error(err, "failed to destroy resize resource once finished", "resize", resize.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

//+kubebuilder:rbac:groups=platform.plural.sh,resources=statefulsetresizes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.plural.sh,resources=statefulsetresizes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.plural.sh,resources=statefulsetresizes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StatefulSetResize object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *StatefulSetResizeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("statefulsetresize", req.NamespacedName)

	var resize platformv1alpha1.StatefulSetResize
	if err := r.Get(ctx, req.NamespacedName, &resize); err != nil {
		log.Error(err, "Failed to fetch resize resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	quant, err := resource.ParseQuantity(resize.Spec.Size)
	if err != nil {
		log.Error(err, "failed to parse size", "size", resize.Spec.Size)
		return ctrl.Result{}, err
	}

	var statefulset appsv1.StatefulSet
	namespacedName := types.NamespacedName{
		Namespace: resize.Namespace,
		Name:      resize.Spec.Name,
	}
	if err := r.Get(ctx, namespacedName, &statefulset); err != nil {
		log.Error(err, "failed to get statefulset", "statefulset", resize.Spec.Name, "namespace", resize.Namespace)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	claims := statefulset.Spec.VolumeClaimTemplates
	newClaims := make([]corev1.PersistentVolumeClaim, len(claims))
	var claim corev1.PersistentVolumeClaim
	for i, cl := range claims {
		if cl.Name == resize.Spec.PersistentVolume {
			if quant.Cmp(cl.Spec.Resources.Requests["storage"]) == 0 {
				log.Info("No change needed for storage")
				return r.cleanup(ctx, &resize)
			}

			cl.Spec.Resources.Requests["storage"] = quant
			claim = cl
		}

		newClaims[i] = cl
	}

	fmt.Printf("%+v", newClaims)

	var storageClass storagev1.StorageClass
	if claim.Spec.StorageClassName == nil {
		classes := &storagev1.StorageClassList{}
		if err := r.List(ctx, classes); err != nil {
			log.Error(err, "could not list storage classes")
			return ctrl.Result{}, err
		}

		found := false
		for _, class := range classes.Items {
			fmt.Printf("%+v", class)
			if _, ok := class.Annotations["storageclass.kubernetes.io/is-default-class"]; ok {
				storageClass = class
				found = true
				break
			}
		}

		if !found {
			err := fmt.Errorf("Could not find default storage class")
			log.Error(err, "could not find default storage class")
			return ctrl.Result{}, err
		}
	} else {
		if err := r.Get(ctx, types.NamespacedName{Name: *claim.Spec.StorageClassName}, &storageClass); err != nil {
			log.Error(err, "failed to get storageClass", "storageclass", *claim.Spec.StorageClassName)
			return ctrl.Result{}, err
		}
	}

	if storageClass.AllowVolumeExpansion == nil || !*storageClass.AllowVolumeExpansion {
		storageClass.AllowVolumeExpansion = resources.BoolPtr(true)
		if err := r.Update(ctx, &storageClass); err != nil {
			log.Error(err, "Failed to enable volume expansion for storage class", "storageclass", storageClass.Name)
			return ctrl.Result{}, err
		}
	}

	replicas := int(*statefulset.Spec.Replicas)

	log.Info("deleting statefulset while orphaning pods", "statefulset", statefulset.Name)
	if err := r.Delete(ctx, &statefulset, client.PropagationPolicy(metav1.DeletePropagationOrphan)); err != nil {
		log.Error(err, "failed to delete", "statefulset", statefulset.Name)
		return ctrl.Result{}, err
	}

	for i := 0; i < replicas; i++ {
		claimName := fmt.Sprintf("%s-%s-%d", claim.Name, statefulset.Name, i)
		var pvc corev1.PersistentVolumeClaim
		nsn := types.NamespacedName{
			Namespace: resize.Namespace,
			Name:      claimName,
		}

		if err := r.Get(ctx, nsn, &pvc); err != nil {
			log.Error(err, "failed to get pvc", "pvc", pvc.Name)
			return ctrl.Result{}, err
		}

		pvc.Spec.Resources.Requests["storage"] = quant

		if err := r.Update(ctx, &pvc); err != nil {
			log.Error(err, "failed to resize statefulset pvc", "pvc", pvc.Name)
			return ctrl.Result{}, err
		}
	}

	if statefulset.Spec.Template.Annotations == nil {
		statefulset.Spec.Template.Annotations = make(map[string]string)
	}

	newStatefulSet := &appsv1.StatefulSet{}
	newStatefulSet.Spec = statefulset.Spec
	newStatefulSet.Spec.VolumeClaimTemplates = newClaims
	newStatefulSet.Name = statefulset.Name
	newStatefulSet.Namespace = statefulset.Namespace
	newStatefulSet.Labels = statefulset.Labels
	newStatefulSet.Annotations = statefulset.Annotations
	if newStatefulSet.Spec.Template.Annotations == nil {
		newStatefulSet.Spec.Template.Annotations = map[string]string{}
	}

	newStatefulSet.Spec.Template.Annotations["platform.plural.sh/rotate"] = time.Now().String()
	if err := r.Create(ctx, newStatefulSet); err != nil {
		log.Error(err, "failed to recreate statefulset", "statefulset", newStatefulSet.Name)
		return ctrl.Result{}, err
	}

	log.Info("Successfully resized pvc for statefulset", "statefulset", newStatefulSet.Name)
	return r.cleanup(ctx, &resize)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StatefulSetResizeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.StatefulSetResize{}).
		Complete(r)
}
