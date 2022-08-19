package redeployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type secretService struct {
	client client.Client
	secret *corev1.Secret
	ctx    context.Context

	workflowMap           map[v1alpha1.WorkflowType]Workflow
	redeployments         []v1alpha1.Redeployment
	matchingRedeployments []v1alpha1.Redeployment
}

func (s *secretService) isControlled(redeployment *v1alpha1.Redeployment) (controlled bool, err error) {
	workflowType := redeployment.Spec.Workflow
	workflow, exists := s.workflowMap[workflowType]

	if !exists {
		workflow, err = newWorkflow(s.client, redeployment)
		if err != nil {
			return
		}

		s.workflowMap[workflowType] = workflow
	}

	if workflow.IsUsed(ResourceSecret, s.secret.Namespace, s.secret.Name) {
		s.matchingRedeployments = append(s.matchingRedeployments, *redeployment)
		controlled = true
	}

	return
}

func (s *secretService) IsControlled() (bool, error) {
	result := false
	for _, redeployment := range s.redeployments {
		controlled, err := s.isControlled(&redeployment)
		if err != nil {
			return false, err
		}

		if controlled {
			result = true
		}
	}

	return result, nil
}

func (s *secretService) HasAnnotation() bool {
	_, ok := s.secret.Annotations[shaAnnotation]
	return ok
}

func (s *secretService) UpdateAnnotation() error {
	sha := s.getSHA()
	s.secret.Annotations[shaAnnotation] = sha

	return s.client.Update(s.ctx, s.secret)
}

func (s *secretService) ShouldRestart() bool {
	existingSHA := s.secret.Annotations[shaAnnotation]
	expectedSHA := s.getSHA()

	return existingSHA != expectedSHA
}

func (s *secretService) RolloutRestart() error {
	for _, redeployment := range s.matchingRedeployments {
		workflow, exists := s.workflowMap[redeployment.Spec.Workflow]
		if !exists {
			return nil
		}

		err := workflow.RolloutRestart(&redeployment)
		if err != nil {
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

func (s *secretService) init() error {
	s.workflowMap = make(map[v1alpha1.WorkflowType]Workflow, 0)
	s.matchingRedeployments = make([]v1alpha1.Redeployment, 0)

	redeployments, err := getRedeployments(s.ctx, s.client, s.secret.Namespace)
	if err != nil {
		return err
	}

	s.redeployments = redeployments
	return nil
}

func newSecretService(client client.Client, secret *corev1.Secret) (Service, error) {
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}

	svc := &secretService{client: client, secret: secret, ctx: context.Background()}

	return svc, svc.init()
}
