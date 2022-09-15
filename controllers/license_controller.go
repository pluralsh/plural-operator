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
	"math/rand"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// LicenseReconciler reconciles a License object
type LicenseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

const (
	pollInterval = 30 * time.Minute
)

const licenseEndpoint = "https://app.plural.sh/api/license/"

//+kubebuilder:rbac:groups=platform.plural.sh,resources=licenses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.plural.sh,resources=licenses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.plural.sh,resources=licenses/finalizers,verbs=update

func fetchLicense(key string) (platformv1alpha1.LicenseStatus, error) {
	var status platformv1alpha1.LicenseStatus
	client := resty.New()

	_, err := client.R().
		SetHeader("Accept", "application/json").
		SetResult(&status).
		Get(licenseEndpoint + key)
	return status, err
}

func (r *LicenseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("license", req.NamespacedName)

	var license platformv1alpha1.License
	if err := r.Get(ctx, req.NamespacedName, &license); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch license resource")
		return ctrl.Result{}, err
	}

	var secret corev1.Secret
	nn := types.NamespacedName{
		Namespace: license.Namespace,
		Name:      license.Spec.SecretRef.Name,
	}
	if err := r.Get(ctx, nn, &secret); err != nil {
		log.Error(err, "could not find secret", "secret", nn.Name)
		return ctrl.Result{}, err
	}

	licenseKey := secret.Data[license.Spec.SecretRef.Key]
	status, err := fetchLicense(string(licenseKey))
	if err != nil {
		log.Error(err, "failed to fetch license")
		return ctrl.Result{}, err
	}

	license.Status = status
	if err := r.Status().Update(ctx, &license); err != nil {
		log.Error(err, "failed to update status of license")
		return ctrl.Result{}, err
	}

	jitter := time.Duration(rand.Intn(60*5)) * time.Second
	return ctrl.Result{RequeueAfter: jitter + pollInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LicenseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.License{}).
		Complete(r)
}
