package alertmanager

import (
	platformv1alpha1 "github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

func removeAlert(alerts []*platformv1alpha1.RunbookAlertStatus, name string) []*platformv1alpha1.RunbookAlertStatus {
	result := make([]*platformv1alpha1.RunbookAlertStatus, 0)
	for _, alert := range alerts {
		if alert.Name != name {
			result = append(result, alert)
		}
	}

	return result
}

func hasAlert(alerts []*platformv1alpha1.RunbookAlertStatus, name string) bool {
	for _, alert := range alerts {
		if alert.Name == name {
			return true
		}
	}

	return false
}

func replaceAlert(alerts []*platformv1alpha1.RunbookAlertStatus, alert *Alert) []*platformv1alpha1.RunbookAlertStatus {
	result := make([]*platformv1alpha1.RunbookAlertStatus, 0)
	name, _ := alert.Labels[nameLabel]
	found := false

	runbookAlert := &platformv1alpha1.RunbookAlertStatus{
		Name:        name,
		StartsAt:    alert.StartsAt,
		Annotations: alert.Annotations,
		Labels:      alert.Labels,
		Fingerprint: alert.Fingerprint,
	}

	for ind, alert := range alerts {
		if alert.Name != name {
			result = append(result, alert)
			continue
		}

		result = append(result, runbookAlert)
		found = true
	}

	if !found {
		result = append(result, runbookAlert)
	}

	return result
}
