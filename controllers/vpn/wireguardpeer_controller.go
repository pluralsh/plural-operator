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
	"strings"

	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	crhelper "github.com/pluralsh/controller-reconcile-helper/pkg"
	"github.com/pluralsh/controller-reconcile-helper/pkg/conditions"
	"github.com/pluralsh/controller-reconcile-helper/pkg/patch"
	crhelperTypes "github.com/pluralsh/controller-reconcile-helper/pkg/types"
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

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

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

	// Always attempt to Patch the Wireguard Server object and status after each reconciliation.
	// This defer block needs to be defined before other code so it is exectued after `ctrl.Result{}` is returned
	defer func() {
		if err := patchWireguardPeer(ctx, patchHelper, peer); err != nil {
			log.Error(err, "failed to patch Wireguard Peer Status")
			utilruntime.HandleError(err)
		}
	}()

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
			conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.FailedToCreateSecretReason, crhelperTypes.ConditionSeverityError, err.Error())
			return ctrl.Result{}, err
		}

		peer.Spec.PublicKey = publicKey
		peer.Spec.PrivateKeyRef = corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret.Name,
			},
			Key: "privateKey",
		}

		return ctrl.Result{Requeue: true}, nil
	}

	wireguard := &vpnv1alpha1.WireguardServer{}
	err = r.Get(ctx, types.NamespacedName{Name: peer.Spec.WireguardRef, Namespace: peer.Namespace}, wireguard)

	if err != nil {
		if errors.IsNotFound(err) {
			conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.WireguardServerNotExistReason, crhelperTypes.ConditionSeverityError, "Wireguard server does not exist")

			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get wireguard")
		return ctrl.Result{}, err
	}

	if !wireguard.Status.Ready {
		log.Info("Wireguard server is not ready")
		conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.WireguardServerNotReadyReason, crhelperTypes.ConditionSeverityError, "Wireguard server is not ready")
		return ctrl.Result{}, nil
	}

	if len(peer.OwnerReferences) == 0 {
		log.Info("Waiting for owner reference to be set " + wireguard.Name + " " + peer.Name)
		if err := ctrl.SetControllerReference(wireguard, peer, r.Scheme); err != nil {
			log.Error(err, "Failed to update peer with controller reference")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	if peer.Status.Config == "" {
		conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.WaitingForConfigReason, crhelperTypes.ConditionSeverityError, "Wireguard peer config has not yet been set")
	} else {
		var config string
		config = peer.Status.Config
		privateKey, err := r.getPrivateKey(ctx, peer)
		if err != nil {
			log.Error(err, "Failed to get private key for generating config")
			conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.FailedToGetPrivateKeyReason, crhelperTypes.ConditionSeverityError, err.Error())
			return ctrl.Result{Requeue: true}, err
		}

		config = strings.ReplaceAll(config, "${PRIVATE_KEY}", privateKey)
		configSecret := r.configSecretForPeer(peer, config)
		log.Info("Creating a new config secret", "secret.Namespace", configSecret.Namespace, "secret.Name", configSecret.Name)
		if err := ctrl.SetControllerReference(peer, configSecret, r.Scheme); err != nil {
			log.Error(err, "Error setting ControllerReference for peer config secret")
			return ctrl.Result{}, err
		}
		if err := crhelper.Secret(ctx, r.Client, configSecret, log); err != nil {
			log.Error(err, "Error reconciling config Secret", "secret", configSecret.Name, "namespace", configSecret.Namespace)
			conditions.MarkFalse(peer, vpnv1alpha1.WireguardPeerReadyCondition, vpnv1alpha1.FailedToCreateSecretReason, crhelperTypes.ConditionSeverityError, err.Error())
			return ctrl.Result{}, err
		}
		peer.Status.ConfigRef = corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: configSecret.Name,
			},
			Key: "wg0.conf",
		}

		conditions.MarkTrue(peer, vpnv1alpha1.WireguardPeerReadyCondition)
	}

	readyCondition := conditions.Get(peer, crhelperTypes.ReadyCondition)
	if readyCondition != nil {
		switch readyCondition.Status {
		case corev1.ConditionFalse, corev1.ConditionUnknown:
			peer.Status.Ready = false
		case corev1.ConditionTrue:
			peer.Status.Ready = true
		}
	}

	return ctrl.Result{}, nil
}

func (r *WireguardPeerReconciler) secretForPeer(m *vpnv1alpha1.WireguardPeer, privateKey string, publicKey string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-peer-keys",
			Namespace: m.Namespace,
			Labels:    labelsForWireguard(m.Name),
		},
		Data: map[string][]byte{"privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}
}

func (r *WireguardPeerReconciler) configSecretForPeer(m *vpnv1alpha1.WireguardPeer, config string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-peer-config",
			Namespace: m.Namespace,
			Labels:    labelsForWireguard(m.Name),
		},
		Data: map[string][]byte{"wg0.conf": []byte(config)},
	}
}

func (r *WireguardPeerReconciler) getPrivateKey(ctx context.Context, peer *vpnv1alpha1.WireguardPeer) (string, error) {
	keySecret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: peer.Spec.PrivateKeyRef.Name, Namespace: peer.Namespace}, keySecret)
	if err != nil {
		return "", err
	}
	if privKey, ok := keySecret.Data[peer.Spec.PrivateKeyRef.Key]; ok {
		return string(privKey), nil
	}
	return "", fmt.Errorf("private key not found in secret")
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardPeerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vpnv1alpha1.WireguardPeer{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

func patchWireguardPeer(ctx context.Context, patchHelper *patch.Helper, wireguardPeer *vpnv1alpha1.WireguardPeer) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it during the deletion process).
	conditions.SetSummary(wireguardPeer,
		conditions.WithConditions(
			vpnv1alpha1.WireguardPeerReadyCondition,
		),
		conditions.WithStepCounterIf(wireguardPeer.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounter(),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		wireguardPeer,
		patch.WithOwnedConditions{Conditions: []crhelperTypes.ConditionType{
			crhelperTypes.ReadyCondition,
			vpnv1alpha1.WireguardPeerReadyCondition,
		},
		},
	)
}
