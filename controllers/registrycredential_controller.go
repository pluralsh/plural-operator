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

	"github.com/go-logr/logr"
	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const dockerconfigjson = ".dockerconfigjson"

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
// the RegistryCredential object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *RegistryCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("RegistryCredential", req.NamespacedName)

	var credentials platformv1alpha1.RegistryCredential
	if err := r.Get(ctx, req.NamespacedName, &credentials); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to fetch registry credentials")
		return ctrl.Result{}, err
	}

	if credentials.DeletionTimestamp != nil {
		log.Info("deleting registry credential secret")
		if err := r.deleteRegistryCredentialSecret(ctx, credentials); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.createOrUpdateRegistryCredentialSecret(ctx, credentials, log); err != nil {
		log.Error(err, "failed to create/update registry credential secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RegistryCredentialsReconciler) createOrUpdateRegistryCredentialSecret(ctx context.Context, credentials platformv1alpha1.RegistryCredential, log logr.Logger) error {
	existingSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: credentials.Namespace, Name: credentials.GetPasswordSecretName()}, existingSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("creating new registry credential secret")
			return r.createSecret(ctx, credentials)
		}
		return err
	}

	return r.updateSecret(ctx, log, credentials, existingSecret)
}

func (r *RegistryCredentialsReconciler) updateSecret(ctx context.Context, log logr.Logger, credentials platformv1alpha1.RegistryCredential, secret *corev1.Secret) error {
	if secret.Data == nil {
		return fmt.Errorf("secret data can not be nil")
	}

	existingAuths := string(secret.Data[dockerconfigjson])
	expectedAuths, err := r.genSecretAuths(ctx, credentials)
	if err != nil {
		return err
	}
	if existingAuths != expectedAuths {
		secret.Data[dockerconfigjson] = []byte(expectedAuths)
		log.Info("updating registry credential secret")
		return r.Update(ctx, secret)
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

	auths, err := r.genSecretAuths(ctx, credentials)
	if err != nil {
		return err
	}
	secret.Data = map[string][]byte{dockerconfigjson: []byte(auths)}

	return r.Create(ctx, secret)
}

func (r *RegistryCredentialsReconciler) genSecretAuths(ctx context.Context, credentials platformv1alpha1.RegistryCredential) (string, error) {
	passwordSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: credentials.Namespace, Name: credentials.Spec.PasswordSecretRef.Name}, passwordSecret); err != nil {
		return "", err
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
		return "", err
	}
	auths := base64.StdEncoding.EncodeToString(rawAuths)
	return auths, nil
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

// requestFromSecret returns a reconcile.Request for the credential registry if the secret is a password reference.
func requestFromSecret(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(mo client.Object) []reconcile.Request {
		secret, ok := mo.(*corev1.Secret)
		if !ok {
			err := fmt.Errorf("object was not a secret but a %T", mo)
			utilruntime.HandleError(err)
			return nil
		}

		registrycredentials := &platformv1alpha1.RegistryCredentialList{}
		if err := c.List(context.Background(), registrycredentials, client.InNamespace(secret.Namespace)); err != nil {
			utilruntime.HandleError(err)
			return nil
		}

		for _, cred := range registrycredentials.Items {
			if cred.Spec.PasswordSecretRef.Name == secret.Name {
				return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: cred.Name, Namespace: cred.Namespace}}}
			}
		}

		return nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.RegistryCredential{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, requestFromSecret(mgr.GetClient())).
		Complete(r)
}
