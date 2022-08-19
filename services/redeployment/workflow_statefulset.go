package redeployment

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type statefulSetWorkflow struct {
	client    client.Client
	namespace string
	ctx       context.Context

	statefulSets *appsv1.StatefulSetList
}

func (d *statefulSetWorkflow) IsUsing(resource Resource, namespace string, name string) bool {
	for _, statefulSet := range d.statefulSets.Items {
		for _, volume := range statefulSet.Spec.Template.Spec.Volumes {
			if isUsing(volume, statefulSet.Namespace, resource, namespace, name) {
				return true
			}
		}
	}

	return false
}

func (d *statefulSetWorkflow) RolloutRestart(redeployment *v1alpha1.Redeployment) error {
	for _, statefulSet := range d.statefulSets.Items {
		if statefulSet.Name == redeployment.Spec.Name && statefulSet.Namespace == redeployment.Spec.Namespace {
			return d.restart(&statefulSet)
		}
	}

	return nil
}

func (d *statefulSetWorkflow) restart(statefulSet *appsv1.StatefulSet) error {
	if statefulSet.Spec.Template.ObjectMeta.Annotations == nil {
		statefulSet.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}

	statefulSet.Spec.Template.ObjectMeta.Annotations[restartAnnotation] = time.Now().Format(time.RFC3339)
	return d.client.Update(d.ctx, statefulSet)
}

func (d *statefulSetWorkflow) init() error {
	return d.client.List(d.ctx, d.statefulSets, &client.ListOptions{Namespace: d.namespace})
}

func newStatefulSetWorkflow(client client.Client, namespace string) (Workflow, error) {
	svc := &statefulSetWorkflow{client: client, namespace: namespace, ctx: context.Background(), statefulSets: &appsv1.StatefulSetList{}}
	return svc, svc.init()
}
