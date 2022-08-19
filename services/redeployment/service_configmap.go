package redeployment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type configMapService struct {
	client    client.Client
	configMap *corev1.ConfigMap
	ctx       context.Context

	workflowMap   map[v1alpha1.WorkflowType]Workflow
	redeployments []v1alpha1.Redeployment
}

// IsControlled implements Service.IsControlled interface.
func (c *configMapService) IsControlled() (bool, error) {
	result := false
	for _, redeployment := range c.redeployments {
		controlled, err := c.isControlled(&redeployment)
		if err != nil {
			return false, err
		}

		if controlled {
			result = true
		}
	}

	return result, nil
}

// HasAnnotation implements Service.HasAnnotation interface.
func (c *configMapService) HasAnnotation() bool {
	_, exists := c.configMap.Annotations[shaAnnotation]
	return exists
}

// UpdateAnnotation implements Service.UpdateAnnotation interface.
func (c *configMapService) UpdateAnnotation() error {
	sha := c.getSHA()
	c.configMap.Annotations[shaAnnotation] = sha

	return c.client.Update(c.ctx, c.configMap)
}

// ShouldRestart implements Service.ShouldRestart interface.
func (c *configMapService) ShouldRestart() bool {
	existingSHA := c.configMap.Annotations[shaAnnotation]
	expectedSHA := c.getSHA()

	return existingSHA != expectedSHA
}

// RolloutRestart implements Service.RolloutRestart interface.
func (c *configMapService) RolloutRestart() error {
	for _, redeployment := range c.redeployments {
		workflow, exists := c.workflowMap[redeployment.Spec.Workflow]
		if !exists {
			return nil
		}

		return workflow.RolloutRestart(&redeployment)
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

func (c *configMapService) isControlled(redeployment *v1alpha1.Redeployment) (controlled bool, err error) {
	workflowType := redeployment.Spec.Workflow
	workflow, exists := c.workflowMap[workflowType]

	if !exists {
		workflow, err = newWorkflow(c.client, redeployment)
		if err != nil {
			return false, err
		}

		c.workflowMap[workflowType] = workflow
	}

	if workflow.IsUsed(ResourceConfigMap, c.configMap.Namespace, c.configMap.Name) {
		c.redeployments = append(c.redeployments, *redeployment)
		controlled = true
	}

	return controlled, err
}

func (c *configMapService) init() error {
	c.workflowMap = make(map[v1alpha1.WorkflowType]Workflow, 0)
	redeployments, err := getRedeployments(c.ctx, c.client, c.configMap.Namespace)
	if err != nil {
		return err
	}

	c.redeployments = redeployments
	return nil
}

func newConfigMapService(client client.Client, configMap *corev1.ConfigMap) (Service, error) {
	if configMap.Annotations == nil {
		configMap.Annotations = map[string]string{}
	}

	svc := &configMapService{client: client, configMap: configMap, ctx: context.Background()}
	return svc, svc.init()
}
