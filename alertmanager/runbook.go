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
