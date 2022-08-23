package redeployment

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

func isUsed(volume corev1.Volume, workflowResourceNamespace string, resource Resource, namespace string, name string) bool {
	switch resource {
	case ResourceConfigMap:
		return volume.ConfigMap != nil && volume.ConfigMap.Name == name && workflowResourceNamespace == namespace
	case ResourceSecret:
		return volume.Secret != nil && volume.Secret.SecretName == name && workflowResourceNamespace == namespace
	}

	return false
}

func isUsedReferance(envFromSource corev1.EnvFromSource, workflowResourceNamespace string, resource Resource, namespace string, name string) bool {
	switch resource {
	case ResourceConfigMap:
		return envFromSource.ConfigMapRef != nil && envFromSource.ConfigMapRef.Name == name && workflowResourceNamespace == namespace
	case ResourceSecret:
		return envFromSource.SecretRef != nil && envFromSource.SecretRef.Name == name && workflowResourceNamespace == namespace
	}

	return false
}

func getRedeployments(ctx context.Context, c client.Client, namespace string) ([]v1alpha1.Redeployment, error) {
	redeploymentList := &v1alpha1.RedeploymentList{}
	result := make([]v1alpha1.Redeployment, 0)
	err := c.List(ctx, redeploymentList, &client.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, r := range redeploymentList.Items {
		if r.Spec.Namespace == namespace {
			result = append(result, r)
		}
	}

	return result, nil
}
