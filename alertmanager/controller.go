package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type AlertmanagerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	route = "/webhook"
)

func (amr *AlertmanagerReconciler) Reconcile(ctx context.Context, payload *WebhookPayload) error {
	log := amr.Log.WithValues("Alertmanager", "reconciler")

	log.Info(fmt.Sprintf("webhook payload %+v", payload))

	runbooks := &platformv1alpha1.RunbookList{}
	if err := amr.List(ctx, runbooks); err != nil {
		return err
	}

	return nil
}

func SetupAlertmanager(ctx context.Context, addr string, amr *AlertmanagerReconciler) error {
	server := http.NewServeMux()

	server.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		log := amr.Log.WithValues("Alertmanager", "handler")
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err, "failed to read alertmanager webhook")
		}

		webhookPayload := WebhookPayload{}
		err = json.Unmarshal([]byte(payload), &webhookPayload)
		if err != nil {
			log.Error(err, "failed to unmarshall alertmanager payload")
		}

		if err := amr.Reconcile(ctx, &webhookPayload); err != nil {
			log.Error(err, "failed to reconcile alertmanager webhook")
		}
	})

	return http.ListenAndServe(addr, server)
}
