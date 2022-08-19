package redeployment

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type secretService struct {
	client client.Client
	secret *corev1.Secret
	ctx    context.Context
}

func (s *secretService) IsControlled() (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) HasAnnotation() bool {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) UpdateAnnotation() error {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) ShouldRestart() bool {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) RolloutRestart() error {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) getSHA() string {
	// TODO implement me
	panic("implement me")
}

func (s *secretService) isControlled(redeployment *v1alpha1.Redeployment) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (c *secretService) init() error {
	//TODO implement me
	panic("implement me")
}

func newSecretService(client client.Client, secret *corev1.Secret) (Service, error) {
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}

	svc := &secretService{client: client, secret: secret, ctx: context.Background()}
	return svc, svc.init()
}
