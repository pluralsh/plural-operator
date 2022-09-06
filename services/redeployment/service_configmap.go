package redeployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	corev1 "k8s.io/api/core/v1"
	ctrclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type configMapService struct {
	client    ctrclient.Client
	configMap *corev1.ConfigMap
	ctx       context.Context

	pods         []corev1.Pod
	matchingPods map[string]corev1.Pod
}

// IsControlled implements Service.IsControlled interface.
func (c *configMapService) IsControlled() bool {
	controlled := false
	for _, pod := range c.pods {
		for _, volume := range pod.Spec.Volumes {
			if isUsed(volume, ResourceConfigMap, c.configMap.Name) {
				controlled = true
				c.matchingPods[pod.Name] = pod
			}
		}
		for _, container := range pod.Spec.Containers {
			for _, secretRef := range container.EnvFrom {
				if isUsedReferance(secretRef, ResourceConfigMap, c.configMap.Name) {
					controlled = true
					c.matchingPods[pod.Name] = pod
				}
			}
		}
	}

	return controlled
}

// HasAnnotation implements Service.HasAnnotation interface.
func (c *configMapService) HasAnnotation() bool {
	_, exists := c.configMap.Annotations[ShaAnnotation]
	return exists
}

// UpdateAnnotation implements Service.UpdateAnnotation interface.
func (c *configMapService) UpdateAnnotation() error {
	sha := c.getSHA()
	c.configMap.Annotations[ShaAnnotation] = sha

	return c.client.Update(c.ctx, c.configMap)
}

// ShouldDeletePods implements Service.ShouldDeletePods interface.
func (c *configMapService) ShouldDeletePods() bool {
	existingSHA := c.configMap.Annotations[ShaAnnotation]
	expectedSHA := c.getSHA()

	return existingSHA != expectedSHA
}

// DeletePods implements Service.DeletePods interface.
func (c *configMapService) DeletePods() error {
	for _, pod := range c.matchingPods {
		if err := c.client.Delete(c.ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

// getSHA implements Service.getSHA interface.
func (c *configMapService) getSHA() string {
	sha := sha256.New()
	dataKeys := make([]string, 0)
	binaryDataKeys := make([]string, 0)

	for key := range c.configMap.Data {
		dataKeys = append(dataKeys, key)
	}

	for key := range c.configMap.BinaryData {
		binaryDataKeys = append(binaryDataKeys, key)
	}

	sort.Strings(dataKeys)
	sort.Strings(binaryDataKeys)

	for _, key := range dataKeys {
		sha.Write([]byte(c.configMap.Data[key]))
	}

	for _, key := range binaryDataKeys {
		sha.Write(c.configMap.BinaryData[key])
	}

	return hex.EncodeToString(sha.Sum(nil))
}

func newConfigMapService(client ctrclient.Client, configMap *corev1.ConfigMap) (Service, error) {
	if configMap.Annotations == nil {
		configMap.Annotations = map[string]string{}
	}
	ctx := context.Background()
	pods := &corev1.PodList{}

	labelSelector, err := RedeployLabelSelector()
	if err != nil {
		return nil, err
	}

	if err := client.List(ctx, pods, &ctrclient.ListOptions{Namespace: configMap.Namespace, LabelSelector: labelSelector}); err != nil {
		return nil, err
	}

	return &configMapService{client: client, configMap: configMap, ctx: ctx, pods: pods.Items, matchingPods: map[string]corev1.Pod{}}, nil
}
