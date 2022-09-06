package redeployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	ctrclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type secretService struct {
	client ctrclient.Client
	secret *corev1.Secret
	ctx    context.Context

	pods         []corev1.Pod
	matchingPods map[string]corev1.Pod
}

func (s *secretService) IsControlled() bool {
	controlled := false
	for _, pod := range s.pods {
		for _, volume := range pod.Spec.Volumes {
			if isUsed(volume, ResourceSecret, s.secret.Name) {
				controlled = true
				s.matchingPods[pod.Name] = pod
			}
		}
		for _, container := range pod.Spec.Containers {
			for _, secretRef := range container.EnvFrom {
				if isUsedReferance(secretRef, ResourceSecret, s.secret.Name) {
					controlled = true
					s.matchingPods[pod.Name] = pod
				}
			}
		}
	}

	return controlled
}

func (s *secretService) HasAnnotation() bool {
	_, ok := s.secret.Annotations[ShaAnnotation]
	return ok
}

func (s *secretService) UpdateAnnotation() error {
	sha := s.getSHA()
	s.secret.Annotations[ShaAnnotation] = sha

	return s.client.Update(s.ctx, s.secret)
}

func (s *secretService) ShouldDeletePods() bool {
	existingSHA := s.secret.Annotations[ShaAnnotation]
	expectedSHA := s.getSHA()

	return existingSHA != expectedSHA
}

func (s *secretService) DeletePods() error {
	for _, pod := range s.matchingPods {
		if err := s.client.Delete(s.ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

func (s *secretService) getSHA() string {
	sha := sha256.New()
	dataKeys := make([]string, 0)

	for key := range s.secret.Data {
		dataKeys = append(dataKeys, key)
	}

	sort.Strings(dataKeys)

	for _, key := range dataKeys {
		sha.Write(s.secret.Data[key])
	}

	return hex.EncodeToString(sha.Sum(nil))
}

func newSecretService(client ctrclient.Client, secret *corev1.Secret) (Service, error) {
	if secret == nil {
		return nil, fmt.Errorf("the secret can not be nil")
	}
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}
	ctx := context.Background()
	pods := &corev1.PodList{}

	labelSelector, err := RedeployLabelSelector()
	if err != nil {
		return nil, err
	}

	if err := client.List(ctx, pods, &ctrclient.ListOptions{Namespace: secret.Namespace, LabelSelector: labelSelector}); err != nil {
		return nil, err
	}

	return &secretService{client: client, secret: secret, ctx: ctx, pods: pods.Items, matchingPods: map[string]corev1.Pod{}}, nil
}
