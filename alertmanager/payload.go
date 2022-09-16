package alertmanager

import (
	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
)

type Alert struct {
	Status      string
	Labels      map[string]string
	Annotations map[string]string
	StartsAt    string
	Fingerprint string
}

type WebhookPayload struct {
	Version  string
	Status   string
	Receiver string
	Alerts   []*Alert
}

const (
	nameLabel      = "alertname"
	ResolvedStatus = "resolved"
	FiringStatus   = "firing"
)

func matchesRunbook(alert *Alert, runbook *platformv1alpha1.Runbook) bool {
	for _, val := range runbook.Spec.Alerts {
		name, ok := alert.Labels[nameLabel]
		namespace, nsOk := alert.Labels["namespace"]

		if ok && name == val.Name && (!nsOk || namespace == runbook.Namespace) {
			return true
		}
	}

	return false
}
