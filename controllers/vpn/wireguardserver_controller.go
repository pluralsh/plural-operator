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
	"strconv"

	"github.com/pluralsh/controller-reconcile-helper/pkg/conditions"
	"github.com/pluralsh/controller-reconcile-helper/pkg/patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	crhelper "github.com/pluralsh/controller-reconcile-helper/pkg"
	crhelperTypes "github.com/pluralsh/controller-reconcile-helper/pkg/types"
	vpnv1alpha1 "github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"
	wgtypes "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Port wireguard runs on
const defaultWireguardPort = int32(51820)

// Port of the wireguard prometheus exporter
const metricsPort = 9586

// WireguardServerReconciler reconciles a Wireguard object
type WireguardServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vpn.plural.sh,resources=wireguards/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wireguard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *WireguardServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("WireguardServer", req.NamespacedName)

	wireguardInstance := &vpnv1alpha1.WireguardServer{}
	if err := r.Get(ctx, req.NamespacedName, wireguardInstance); err != nil {
		if apierrs.IsNotFound(err) {
			// log.Info("Unable to fetch wireguard server - skipping", "namespace", wireguardInstance.Namespace, "name", wireguardInstance.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	//if wireguard is terminating , no need to reconcile
	if wireguardInstance.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	var wireguardPort int32
	wireguardPort = defaultWireguardPort

	if wireguardInstance.Spec.Port != nil {
		wireguardPort = *wireguardInstance.Spec.Port
	}

	// The PatchHelper needs to be instantiated before the status is changed
	// this is because it performs a diff between the instantiated status and
	// the object passed in the Path function
	patchHelper, err := patch.NewHelper(wireguardInstance, r.Client)
	if err != nil {
		log.Error(err, "Error getting patchHelper for WireguardServer")
		return ctrl.Result{}, err
	}

	// Reconcile Service for the wireguard server
	service := r.generateService(wireguardInstance, wireguardPort)
	if err := ctrl.SetControllerReference(wireguardInstance, service, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for Service")
		return ctrl.Result{}, err
	}
	if err := crhelper.Service(ctx, r.Client, service, log); err != nil {
		log.Error(err, "Error reconciling Service", "service", service.Name, "namespace", service.Namespace)
		conditions.MarkFalse(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition, vpnv1alpha1.FailedToCreateServiceReason, crhelperTypes.ConditionSeverityError, err.Error())
		return ctrl.Result{}, err
	} else {
		conditions.MarkTrue(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition)
	}

	// Reconcile Service for the wireguard server
	metricsService := r.generateMetricsService(wireguardInstance)
	if err := ctrl.SetControllerReference(wireguardInstance, metricsService, r.Scheme); err != nil {
		log.Error(err, "Error setting ControllerReference for Metrics Service")
		return ctrl.Result{}, err
	}
	if err := crhelper.Service(ctx, r.Client, metricsService, log); err != nil {
		log.Error(err, "Error reconciling Metrics Service", "service", metricsService.Name, "namespace", metricsService.Namespace)
		conditions.MarkFalse(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition, vpnv1alpha1.FailedToCreateMetricsServiceReason, crhelperTypes.ConditionSeverityError, err.Error())
		return ctrl.Result{}, err
	}

	foundSecret := &corev1.Secret{}

	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundSecret); err == nil {
		privateKey := string(foundSecret.Data["privateKey"])
		publicKey := string(foundSecret.Data["publicKey"])
		wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.8.0.1/24
ListenPort = %v
`, privateKey, wireguardPort)

		peerConfig := ""

		// wireguardpeer
		peers := &vpnv1alpha1.WireguardPeerList{}
		if err := r.List(ctx, peers, client.InNamespace(req.Namespace)); err != nil {
			log.Error(err, "Failed to fetch list of peers")
			return ctrl.Result{}, err
		}

		for _, peer := range peers.Items {

			if peer.Spec.WireguardRef != wireguardInstance.Name {
				continue
			}
			// if peer.Spec.PublicKey == "" {
			// 	continue
			// }

			peerConfig = peerConfig + fmt.Sprintf("\n[Peer]\nPublicKey = %s\nallowedIps = %s\n\n", peer.Spec.PublicKey, peer.Spec.Address)
		}

		wgConfig = wgConfig + peerConfig

		log.Info("secret", "wgconfig", wgConfig)

		secret := r.generateSecret(wireguardInstance, privateKey, publicKey, wgConfig)
		if err := ctrl.SetControllerReference(wireguardInstance, secret, r.Scheme); err != nil {
			log.Error(err, "Error setting ControllerReference for Secret")
			return ctrl.Result{}, err
		}
		if err := crhelper.Secret(ctx, r.Client, secret, log); err != nil {
			log.Error(err, "Error reconciling Secret", "secret", secret.Name, "namespace", secret.Namespace)
			conditions.MarkFalse(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition, vpnv1alpha1.FailedToCreateSecretReason, crhelperTypes.ConditionSeverityError, err.Error())
			return ctrl.Result{}, err
		}

	} else {
		if apierrs.IsNotFound(err) {

			key, err := wgtypes.GeneratePrivateKey()
			if err != nil {
				log.Error(err, "Error generating wireguard key")
				return ctrl.Result{}, err
			}

			privateKey := key.String()
			publicKey := key.PublicKey().String()

			wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = 10.8.0.1/24
ListenPort = %v
`, privateKey, wireguardPort)

			secret := r.generateSecret(wireguardInstance, privateKey, publicKey, wgConfig)
			if err := ctrl.SetControllerReference(wireguardInstance, secret, r.Scheme); err != nil {
				log.Error(err, "Error setting ControllerReference for Secret")
				return ctrl.Result{}, err
			}
			if err := crhelper.Secret(ctx, r.Client, secret, log); err != nil {
				log.Error(err, "Error reconciling Secret", "secret", secret.Name, "namespace", secret.Namespace)
				conditions.MarkFalse(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition, vpnv1alpha1.FailedToCreateSecretReason, crhelperTypes.ConditionSeverityError, err.Error())
				return ctrl.Result{}, err
			}

			// return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	foundService := &corev1.Service{}

	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	serviceType := foundService.Spec.Type

	hostname := ""
	port := defaultWireguardPort

	if serviceType == corev1.ServiceTypeLoadBalancer {
		ingressList := foundService.Status.LoadBalancer.Ingress
		log.Info("Found ingress", "ingress", ingressList)
		if len(ingressList) == 0 {
			conditions.MarkFalse(wireguardInstance, vpnv1alpha1.WireguardServerReadyCondition, vpnv1alpha1.ServiceNotReadyReason, crhelperTypes.ConditionSeverityInfo, "")
			return ctrl.Result{}, nil
		}

		hostname = foundService.Status.LoadBalancer.Ingress[0].Hostname

		if hostname == "" {
			hostname = foundService.Status.LoadBalancer.Ingress[0].IP
		}
		port = foundService.Spec.Ports[0].Port
	}

	wireguardInstance.Status.Hostname = hostname
	wireguardInstance.Status.Port = strconv.Itoa(int(port))

	readyCondition := conditions.Get(wireguardInstance, crhelperTypes.ReadyCondition)
	if readyCondition != nil {
		switch readyCondition.Status {
		case corev1.ConditionFalse, corev1.ConditionUnknown:
			wireguardInstance.Status.Ready = false
		case corev1.ConditionTrue:
			wireguardInstance.Status.Ready = true
		}
	}

	// Always attempt to Patch the Wireguard Server object and status after each reconciliation.
	defer func() {

		patchWireguardServer(ctx, patchHelper, wireguardInstance)
		if err := patchWireguardServer(ctx, patchHelper, wireguardInstance); err != nil {
			log.Error(err, "failed to patch Wireguard Server Status")
			// if rerr == nil {
			// 	rerr = err
			// }
		}
	}()

	// return r.reconcile(ctx, wireguardInstance, log)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WireguardServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vpnv1alpha1.WireguardServer{}).
		// Owns(&vpnv1alpha1.WireguardPeer{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Watches(&source.Kind{Type: &vpnv1alpha1.WireguardPeer{}},
			handler.EnqueueRequestsFromMapFunc(
				func(obj client.Object) []reconcile.Request {
					req := []reconcile.Request{}
					serverList := &vpnv1alpha1.WireguardServerList{}
					err := r.List(context.TODO(), serverList)
					if err != nil {
						r.Log.Error(err, "Failed to list WireguardServers in order to trigger reconciliation")
						return req
					}
					for _, server := range serverList.Items {
						req = append(req, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      server.Name,
								Namespace: server.Namespace,
							}})
					}
					return req
				})).
		Complete(r)
}

// Generate the desired Service object for the workspace
func (r *WireguardServerReconciler) generateService(wireguardInstance *vpnv1alpha1.WireguardServer, wireguardPort int32) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wireguardInstance.Name,
			Namespace: wireguardInstance.Namespace,
			Labels: map[string]string{
				"wireguardserver.vpn.plural.sh":             "",
				"wireguardserver.vpn.plural.sh/server-name": wireguardInstance.Name,
			},
			//TODO: make this configurable
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type":            "external",
				"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": "ip",
				"service.beta.kubernetes.io/aws-load-balancer-scheme":          "internet-facing",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: wireguardInstance.Spec.ServiceType,
			Selector: map[string]string{
				"wireguardserver.vpn.plural.sh":             "",
				"wireguardserver.vpn.plural.sh/server-name": wireguardInstance.Name,
			},
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolUDP,
				Port:       wireguardPort,
				TargetPort: intstr.FromInt(int(wireguardPort)),
			}},
		},
	}
	return svc
}

func (r *WireguardServerReconciler) generateMetricsService(wireguardInstance *vpnv1alpha1.WireguardServer) *corev1.Service {

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wireguardInstance.Name + "-metrics",
			Namespace: wireguardInstance.Namespace,
			Labels: map[string]string{
				"wireguardserver.vpn.plural.sh":             "",
				"wireguardserver.vpn.plural.sh/server-name": wireguardInstance.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"wireguardserver.vpn.plural.sh":             "",
				"wireguardserver.vpn.plural.sh/server-name": wireguardInstance.Name,
			},
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       metricsPort,
				TargetPort: intstr.FromInt(metricsPort),
			}},
		},
	}
	return svc
}

func (r *WireguardServerReconciler) generateSecret(wireguardInstance *vpnv1alpha1.WireguardServer, privateKey string, publicKey string, config string) *corev1.Secret {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wireguardInstance.Name,
			Namespace: wireguardInstance.Namespace,
			Labels: map[string]string{
				"wireguardserver.vpn.plural.sh":             "",
				"wireguardserver.vpn.plural.sh/server-name": wireguardInstance.Name,
			},
		},
		Data: map[string][]byte{"config": []byte(config), "privateKey": []byte(privateKey), "publicKey": []byte(publicKey)},
	}
	return secret
}

func patchWireguardServer(ctx context.Context, patchHelper *patch.Helper, wireguardServer *vpnv1alpha1.WireguardServer) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding it during the deletion process).
	conditions.SetSummary(wireguardServer,
		conditions.WithConditions(
			vpnv1alpha1.WireguardServerReadyCondition,
		),
		conditions.WithStepCounterIf(wireguardServer.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounter(),
	)

	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	return patchHelper.Patch(
		ctx,
		wireguardServer,
		patch.WithOwnedConditions{Conditions: []crhelperTypes.ConditionType{
			crhelperTypes.ReadyCondition,
			vpnv1alpha1.WireguardServerReadyCondition,
		},
		},
	)
}
