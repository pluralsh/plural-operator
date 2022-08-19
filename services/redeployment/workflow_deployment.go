package redeployment

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type deploymentWorkflow struct {
	client    client.Client
	namespace string
	ctx       context.Context

	deployments *appsv1.DeploymentList
}

func (d *deploymentWorkflow) IsUsed(resource Resource, namespace string, name string) bool {
	for _, deployment := range d.deployments.Items {
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if isUsed(volume, deployment.Namespace, resource, namespace, name) {
				return true
			}
		}
	}

	return false
}

func (d *deploymentWorkflow) RolloutRestart(redeployment *v1alpha1.Redeployment) error {
	for _, deployment := range d.deployments.Items {
		if deployment.Name == redeployment.Spec.Name && deployment.Namespace == redeployment.Spec.Namespace {
			return d.restart(deployment)
		}
	}

	return nil
}

func (d *deploymentWorkflow) restart(deployment appsv1.Deployment) error {
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}

	deployment.Spec.Template.ObjectMeta.Annotations[restartAnnotation] = time.Now().Format(time.RFC3339)
	return d.client.Update(d.ctx, &deployment)
}

func (d *deploymentWorkflow) init() error {
	return d.client.List(d.ctx, d.deployments, &client.ListOptions{Namespace: d.namespace})
}

func newDeploymentWorkflow(client client.Client, namespace string) (Workflow, error) {
	svc := &deploymentWorkflow{client: client, namespace: namespace, ctx: context.Background(), deployments: &appsv1.DeploymentList{}}
	return svc, svc.init()
}
