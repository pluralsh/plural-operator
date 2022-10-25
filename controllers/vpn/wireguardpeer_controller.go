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

package vpn

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	crhelper "github.com/pluralsh/controller-reconcile-helper/pkg"
	"github.com/pluralsh/controller-reconcile-helper/pkg/patch"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	vpnv1alpha1 "github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"
)

// WireguardPeerReconciler reconciles a WireguardPeer object
type WireguardPeerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguardpeers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguardpeers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguardpeers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WireguardPeer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *WireguardPeerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("WireguardPeer", req.NamespacedName)
	peer := &vpnv1alpha1.WireguardPeer{}
	err := r.Get(ctx, req.NamespacedName, peer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("wireguard peer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get wireguard peer")
		return ctrl.Result{}, err
	}

	// The PatchHelper needs to be instantiated before the status is changed
	// this is because it performs a diff between the instantiated status and
	// the object passed in the Path function
	patchHelper, err := patch.NewHelper(peer, r.Client)
	if err != nil {
		log.Error(err, "Error getting patchHelper for WireguardPeer")
		return ctrl.Result{}, err
	}

	newPeer := peer.DeepCopy()
	if newPeer.Status.Status == "" {
		newPeer.Status.Status = vpnv1alpha1.Pending
		newPeer.Status.Message = "Waiting for wireguard peer to be created"
		if err := patchHelper.Patch(ctx, newPeer); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if peer.Spec.PublicKey == "" {
		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			log.Error(err, "Failed to generate private key")
			return ctrl.Result{}, err
		}

		privateKey := key.String()
		publicKey := key.PublicKey().String()

		secret := r.secretForPeer(peer, privateKey, publicKey)
		log.Info("Creating a new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		if err := ctrl.SetControllerReference(peer, secret, r.Scheme); err != nil {
			log.Error(err, "Error setting ControllerReference for peer secret")
			return ctrl.Result{}, err
		}
		if err := crhelper.Secret(ctx, r.Client, secret, log); err != nil {
			log.Error(err, "Error reconciling Secret", "secret", secret.Name, "namespace", secret.Namespace)
			return ctrl.Result{}, err
		}

		newPeer.Spec.PublicKey = publicKey
		newPeer.Spec.PrivateKey = vpnv1alpha1.PrivateKey{
			SecretKeyRef: corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: peer.Name + "-peer"}, Key: "privateKey"},
		}

		if err := patchHelper.Patch(ctx, newPeer); err != nil {
			log.Error(err, "Failed to create new peer", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	wireguard := &vpnv1alpha1.WireguardServer{}
	err = r.Get(ctx, types.NamespacedName{Name: newPeer.Spec.WireguardRef, Namespace: newPeer.Namespace}, wireguard)

	if err != nil {
		if errors.IsNotFound(err) {
			newPeer.Status.Status = vpnv1alpha1.Error
			newPeer.Status.Message = fmt.Sprintf("Waiting for wireguard resource '%s' to be created", newPeer.Spec.WireguardRef)
			if err := patchHelper.Patch(ctx, newPeer); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get wireguard")

		return ctrl.Result{}, err

	}

	if !wireguard.Status.Ready {
		log.Info("Waiting for wireguard to be ready")
		newPeer.Status.Status = vpnv1alpha1.Error
		newPeer.Status.Message = fmt.Sprintf("Waiting for %s to be ready", wireguard.Name)
		if err := patchHelper.Patch(ctx, newPeer); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	wireguardSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: newPeer.Spec.WireguardRef, Namespace: newPeer.Namespace}, wireguardSecret)

	if len(newPeer.OwnerReferences) == 0 {
		log.Info("Waiting for owner reference to be set " + wireguard.Name + " " + newPeer.Name)
		if err := ctrl.SetControllerReference(wireguard, newPeer, r.Scheme); err != nil {
			log.Error(err, "Failed to update peer with controller reference")
			return ctrl.Result{}, err
		}
		if err := patchHelper.Patch(ctx, newPeer); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if newPeer.Status.Config == "" {
		newPeer.Status.Status = vpnv1alpha1.Pending
		newPeer.Status.Message = "Waiting config to be updated"
		if err := patchHelper.Patch(ctx, newPeer); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *WireguardPeerReconciler) secretForPeer(m *vpnv1alpha1.WireguardPeer, privateKey string, publicKey string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-peer",
			Namespace: m.Namespace,
			Labels:    labelsForWireguard(m.Name),
		},
		Data: map[string][]byte{"privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vpnv1alpha1.WireguardPeer{}).
		Complete(r)
}