package redeployment

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type defaultWorkflowFactory struct{}

func (w *defaultWorkflowFactory) Create(client client.Client, redeployment *v1alpha1.Redeployment) (Workflow, error) {
	switch redeployment.Spec.Workflow {
	case v1alpha1.Deployment:
		return newDeploymentWorkflow(client, redeployment.Spec.Namespace)
	case v1alpha1.StatefulSet:
		return newStatefulSetWorkflow(client, redeployment.Spec.Namespace)
	case v1alpha1.DaemonSet:
		return newDaemonSetWorkflow(client, redeployment.Spec.Namespace)
	}

	panic(fmt.Sprintf("unsupported workflow type found: %s", redeployment.Spec.Workflow))
}

func newWorkflowFactory() WorkflowFactory {
	return &defaultWorkflowFactory{}
}
