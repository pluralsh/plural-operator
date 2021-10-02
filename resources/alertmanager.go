package resources

import (
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AlertManagerConfig(name, url string, labels map[string]string) *v1alpha1.AlertmanagerConfig {
	return &v1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.AlertmanagerConfigSpec{
			Route: &v1alpha1.Route{
				Receiver: name,
				Matchers: buildMatchers(labels),
			},
			Receivers: []v1alpha1.Receiver{
				{
					Name: name,
					WebhookConfigs: []v1alpha1.WebhookConfig{
						{SendResolved: boolPtr(true), URL: stringPtr(url), MaxAlerts: 0},
					},
				},
			},
		},
	}
}

func buildMatchers(labels map[string]string) []v1alpha1.Matcher {
	result := make([]v1alpha1.Matcher, 0)

	for k, v := range labels {
		result = append(result, v1alpha1.Matcher{
			Name:  k,
			Value: v,
		})
	}

	return result
}
