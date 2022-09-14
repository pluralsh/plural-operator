package redeployment

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

// RequestSecretFromPod returns a reconcile.Request for the Pod with RedeployLabel in order to set ShaAnnotation in the secrets.
func RequestSecretFromPod(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(mo client.Object) []reconcile.Request {
		pod, ok := mo.(*corev1.Pod)
		if !ok {
			err := fmt.Errorf("object was not a pod but a %T", mo)
			utilruntime.HandleError(err)
			return nil
		}

		if pod.Labels == nil {
			return nil
		}
		if pod.Labels[RedeployLabel] != "true" {
			return nil
		}

		var result []reconcile.Request

		for _, volume := range pod.Spec.Volumes {
			if volume.Secret != nil {
				result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: volume.Secret.SecretName, Namespace: pod.Namespace}})
			}
		}
		for _, container := range pod.Spec.Containers {
			for _, envRef := range container.EnvFrom {
				if envRef.SecretRef != nil {
					result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: envRef.SecretRef.Name, Namespace: pod.Namespace}})
				}
			}
		}

		return result
	})
}

// RequestConfigMapFromPod returns a reconcile.Request for the Pod with RedeployLabel in order to set ShaAnnotation in the config map.
func RequestConfigMapFromPod(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(mo client.Object) []reconcile.Request {
		pod, ok := mo.(*corev1.Pod)
		if !ok {
			err := fmt.Errorf("object was not a pod but a %T", mo)
			utilruntime.HandleError(err)
			return nil
		}

		if pod.Labels == nil {
			return nil
		}
		if pod.Labels[RedeployLabel] != "true" {
			return nil
		}

		var result []reconcile.Request

		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: volume.ConfigMap.Name, Namespace: pod.Namespace}})
			}
		}
		for _, container := range pod.Spec.Containers {
			for _, envRef := range container.EnvFrom {
				if envRef.ConfigMapRef != nil {
					result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{Name: envRef.ConfigMapRef.Name, Namespace: pod.Namespace}})
				}
			}
		}

		return result
	})
}
