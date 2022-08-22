package redeployment

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type daemonSetWorkflow struct {
	client    client.Client
	namespace string
	ctx       context.Context

	daemonSets *appsv1.DaemonSetList
}

func (d *daemonSetWorkflow) IsUsed(resource Resource, namespace string, name string) bool {
	for _, daemonSet := range d.daemonSets.Items {
		for _, volume := range daemonSet.Spec.Template.Spec.Volumes {
			if isUsed(volume, daemonSet.Namespace, resource, namespace, name) {
				return true
			}
		}
	}

	return false
}

func (d *daemonSetWorkflow) RolloutRestart(redeployment *v1alpha1.Redeployment) error {
	for _, daemonSet := range d.daemonSets.Items {
		if daemonSet.Name == redeployment.Spec.Name && daemonSet.Namespace == redeployment.Spec.Namespace {
			return d.restart(&daemonSet)
		}
	}

	return nil
}

func (d *daemonSetWorkflow) restart(daemonSet *appsv1.DaemonSet) error {
	if daemonSet.Spec.Template.ObjectMeta.Annotations == nil {
		daemonSet.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}

	daemonSet.Spec.Template.ObjectMeta.Annotations[RestartAnnotation] = time.Now().Format(time.RFC3339)
	return d.client.Update(d.ctx, daemonSet)
}

func (d *daemonSetWorkflow) init() error {
	return d.client.List(d.ctx, d.daemonSets, &client.ListOptions{Namespace: d.namespace})
}

func newDaemonSetWorkflow(client client.Client, namespace string) (Workflow, error) {
	svc := &daemonSetWorkflow{client: client, namespace: namespace, ctx: context.Background(), daemonSets: &appsv1.DaemonSetList{}}
	return svc, svc.init()
}
