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

package main

import (
	"crypto/sha256"
	"flag"
	"io/ioutil"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pluralsh/plural-operator/alertmanager"
	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural-operator/controllers"

	amv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/webhook"

	securityv1alpha1 "github.com/pluralsh/plural-operator/api/security/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(platformv1alpha1.AddToScheme(scheme))

	utilruntime.Must(amv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func loadConfig(configFile string) (*securityv1alpha1.Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	klog.Infof("New configuration: sha256sum %x", sha256.Sum256(data))

	var cfg securityv1alpha1.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	// var webhookAddr string
	var oauthSidecarConfig string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	// flag.StringVar(&webhookAddr, "webhook-bind-address", ":3000", "The address the webhook endpoint binds to.")
	flag.StringVar(&oauthSidecarConfig, "oauth-sidecar-config-path", "/tmp/k8s-webhook-server/config/oauth-sidecar-config.yaml", "OAuth Webhook sidecar config")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "0247ec41.plural.sh",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.SecretSyncReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SecretSync"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretSync")
		os.Exit(1)
	}

	if err = (&controllers.ServiceAccountReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ServiceAccount"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceAccount")
		os.Exit(1)
	}

	if err = (&controllers.JobReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Job"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Job")
		os.Exit(1)
	}

	if err = (&controllers.NamespaceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Namespace"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	amr := &alertmanager.AlertmanagerReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Alertmanager"),
		Scheme: mgr.GetScheme(),
	}

	if err := amr.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "alertmanager")
		os.Exit(1)
	}

	// Setup oauth injector mutating webhook
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		config, err := loadConfig(oauthSidecarConfig)
		if err != nil {
			setupLog.Error(err, "unable to load oauth injector config")
			os.Exit(1)
		}

		mgr.GetWebhookServer().Register(
			"/mutate-security-plural-sh-v1alpha1-oauthinjector",
			&webhook.Admission{
				Handler: &securityv1alpha1.OAuthInjector{
					Name:          "oauth2-proxy",
					Log:           ctrl.Log.WithName("webhooks").WithName("oauth-injector"),
					Client:        mgr.GetClient(),
					SidecarConfig: config,
				},
			},
		)
	}

	if err = (&controllers.StatefulSetResizeReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("StatefulSetResize"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StatefulSetResize")
		os.Exit(1)
	}
	if err = (&controllers.DefaultStorageClassReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("DefaultStorageClass"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DefaultStorageClass")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// add webhook handler for alertmanager
	ctx := ctrl.SetupSignalHandler()
	if err := mgr.AddMetricsExtraHandler("/webhook", alertmanager.AlertmanagerHandler(ctx, amr)); err != nil {
		setupLog.Error(err, "unable to set up alertmanager webhook")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
