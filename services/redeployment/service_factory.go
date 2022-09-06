package redeployment

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewService(resource Resource, client client.Client, object client.Object) (Service, error) {
	switch resource {
	case ResourceConfigMap:
		configMap := object.(*corev1.ConfigMap)
		return newConfigMapService(client, configMap)
	case ResourceSecret:
		secret := object.(*corev1.Secret)
		return newSecretService(client, secret)
	}

	panic(fmt.Sprintf("unsupported resource found: %s", resource))
}
