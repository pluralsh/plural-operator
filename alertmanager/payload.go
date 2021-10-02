package alertmanager

import (
	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
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
		if name, ok := alert.Labels[nameLabel]; ok && name == val.Name {
			return true
		}
	}

	return false
}
