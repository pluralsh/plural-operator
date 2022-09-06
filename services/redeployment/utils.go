package redeployment

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func isUsed(volume corev1.Volume, resource Resource, name string) bool {
	switch resource {
	case ResourceConfigMap:
		return volume.ConfigMap != nil && volume.ConfigMap.Name == name
	case ResourceSecret:
		return volume.Secret != nil && volume.Secret.SecretName == name
	}

	return false
}

func isUsedReferance(envFromSource corev1.EnvFromSource, resource Resource, name string) bool {
	switch resource {
	case ResourceConfigMap:
		return envFromSource.ConfigMapRef != nil && envFromSource.ConfigMapRef.Name == name
	case ResourceSecret:
		return envFromSource.SecretRef != nil && envFromSource.SecretRef.Name == name
	}

	return false
}

func RedeployLabelSelector() (labels.Selector, error) {
	req, err := labels.NewRequirement(RedeployLabel, selection.Equals, []string{"true"})
	if err != nil {
		return nil, fmt.Errorf("failed to build label selector: %w", err)
	}
	return labels.Parse(req.String())
}
