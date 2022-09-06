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
	"encoding/base64"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type RegistryCred struct {
	Auth     string `json:"auth"`
	Password string `json:"password"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Auths struct {
	Cred map[string]RegistryCred `json:"auths"`
}

// RegistryCredentialsReconciler reconciles a RegistryCredential object
type RegistryCredentialsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretSync object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *RegistryCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("RegistryCredential", req.NamespacedName)

	var credentials platformv1alpha1.RegistryCredential
	if err := r.Get(ctx, req.NamespacedName, &credentials); err != nil {
		log.Error(err, "failed to fetch registry credentials")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if credentials.DeletionTimestamp != nil {
		log.Info("deleting registry credential secret")
		if err := r.deleteRegistryCredentialSecret(ctx, credentials); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.createRegistryCredentialSecret(ctx, credentials); err != nil {
		log.Error(err, "failed to create registry credential secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RegistryCredentialsReconciler) createRegistryCredentialSecret(ctx context.Context, credentials platformv1alpha1.RegistryCredential) error {
	if err := r.Get(ctx, client.ObjectKey{Namespace: credentials.Namespace, Name: credentials.GetPasswordSecretName()}, &corev1.Secret{}); err != nil {
		if apierrors.IsNotFound(err) {
			return r.createSecret(ctx, credentials)
		}
		return err
	}

	return nil
}

func (r *RegistryCredentialsReconciler) createSecret(ctx context.Context, credentials platformv1alpha1.RegistryCredential) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentials.GetPasswordSecretName(),
			Namespace: credentials.Namespace,
			Labels:    map[string]string{secretSyncLabel: "all", secretTypeLabel: "plural-creds"},
		},
		Type: "kubernetes.io/dockerconfigjson",
	}

	passwordSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: credentials.Namespace, Name: credentials.Spec.PasswordSecretRef.Name}, passwordSecret); err != nil {
		return err
	}
	password := passwordSecret.Data[credentials.Spec.PasswordSecretRef.Key]

	rawAuth := fmt.Sprintf("%s:%s", credentials.Spec.Username, password)
	auth := base64.StdEncoding.EncodeToString([]byte(rawAuth))

	a := Auths{
		Cred: map[string]RegistryCred{credentials.Spec.Server: {
			Auth:     auth,
			Password: string(password),
			Username: credentials.Spec.Username,
			Email:    credentials.Spec.Email,
		}},
	}

	rawAuths, err := json.Marshal(a)
	if err != nil {
		return err
	}
	auths := base64.StdEncoding.EncodeToString(rawAuths)
	secret.Data = map[string][]byte{".dockerconfigjson": []byte(auths)}

	return r.Create(ctx, secret)
}

func (r *RegistryCredentialsReconciler) deleteRegistryCredentialSecret(ctx context.Context, credentials platformv1alpha1.RegistryCredential) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentials.GetPasswordSecretName(),
			Namespace: credentials.Namespace,
		},
	}
	return r.Delete(ctx, secret)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.RegistryCredential{}).
		Complete(r)
}
