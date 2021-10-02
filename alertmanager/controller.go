package alertmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
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

func (amr *AlertmanagerReconciler) HandleWebhook(ctx context.Context, payload *WebhookPayload) error {
	log := amr.Log.WithValues("Alertmanager", "reconciler")

	log.Info(fmt.Sprintf("webhook payload %+v", payload))

	runbooks := &platformv1alpha1.RunbookList{}
	if err := amr.List(ctx, runbooks); err != nil {
		return err
	}

	for _, runbook := range runbooks.Items {
		alerts := runbook.Status.Alerts
		hasMatch := false
		for _, alert := range payload.Alerts {
			if !matchesRunbook(alert, &runbook) {
				continue
			}
			hasMatch = true

			name, _ := alert.Labels[nameLabel]
			if alert.Status == ResolvedStatus {
				alerts = removeAlert(alerts, name)
			} else if !hasAlert(alerts, name) {
				alerts = append(alerts, &platformv1alpha1.RunbookAlertStatus{
					Name:        name,
					StartsAt:    alert.StartsAt,
					Annotations: alert.Annotations,
					Labels:      alert.Labels,
					Fingerprint: alert.Fingerprint,
				})
			}
		}

		if hasMatch {
			runbook.Status.Alerts = alerts
			if err := amr.Status().Update(ctx, &runbook); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *AlertmanagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// no-op
	return ctrl.Result{}, nil
}

func (r *AlertmanagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.Runbook{}).
		Complete(r)
}

func AlertmanagerHandler(ctx context.Context, amr *AlertmanagerReconciler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := amr.HandleWebhook(ctx, &webhookPayload); err != nil {
			log.Error(err, "failed to reconcile alertmanager webhook")
		}
	}
}
