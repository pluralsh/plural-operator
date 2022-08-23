package redeployment

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type podWorkflow struct {
	client    client.Client
	namespace string
	ctx       context.Context

	pods *corev1.PodList
}

func (d *podWorkflow) IsUsed(resource Resource, namespace string, name string) bool {
	for _, pod := range d.pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if isUsed(volume, pod.Namespace, resource, namespace, name) {
				return true
			}
		}
		for _, container := range pod.Spec.Containers {
			for _, secretRef := range container.EnvFrom {
				if isUsedReferance(secretRef, pod.Namespace, resource, namespace, name) {
					return true
				}
			}
		}
	}

	return false
}

func (d *podWorkflow) RolloutRestart(redeployment *v1alpha1.Redeployment) error {
	for _, pod := range d.pods.Items {
		if pod.Name == redeployment.Spec.Name && pod.Namespace == redeployment.Spec.Namespace {
			return d.restart(pod)
		}
	}

	return nil
}

func (d *podWorkflow) restart(pod corev1.Pod) error {
	var gracePeriodSeconds int64
	gracePeriodSeconds = 0
	copy := pod.DeepCopy()
	copy.ObjectMeta = metav1.ObjectMeta{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}
	copy.Status = corev1.PodStatus{}

	if err := d.client.Delete(d.ctx, &pod, &client.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}); err != nil {
		return err
	}

	return d.client.Create(d.ctx, copy)
}

func (d *podWorkflow) init() error {
	return d.client.List(d.ctx, d.pods, &client.ListOptions{Namespace: d.namespace})
}

func newPodWorkflow(client client.Client, namespace string) (Workflow, error) {
	svc := &podWorkflow{client: client, namespace: namespace, ctx: context.Background(), pods: &corev1.PodList{}}
	return svc, svc.init()
}
