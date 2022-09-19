package redeployment

import (
	"context"
	"fmt"
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
		if pod.Annotations == nil {
			continue
		}
		if pod.Annotations["security.plural.sh/oauth-env-secret"] == s.secret.Name || pod.Annotations["security.plural.sh/htpasswd-secret"] == s.secret.Name {
			controlled = true
			s.matchingPods[pod.Name] = pod
		}
	}

	return controlled
}

func (s *secretService) DeletePods() error {
	for _, pod := range s.matchingPods {
		if err := s.client.Delete(s.ctx, &pod); err != nil {
			return err
		}
	}

	return nil
}

func NewSecretService(client ctrclient.Client, secret *corev1.Secret) (Service, error) {
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
