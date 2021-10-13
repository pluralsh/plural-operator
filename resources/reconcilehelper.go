package resources

import (
	"context"
	"reflect"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-logr/logr"

	// istioNetworking "istio.io/api/networking/v1beta1"

	// istioSecurity "istio.io/api/security/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment reconciles a k8s deployment object.
func Deployment(ctx context.Context, r client.Client, deployment *appsv1.Deployment, log logr.Logger) error {
	foundDeployment := &appsv1.Deployment{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
			if err := r.Create(ctx, deployment); err != nil {
				log.Error(err, "unable to create deployment")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "error getting deployment")
			return err
		}
	}
	if !justCreated && CopyDeploymentSetFields(deployment, foundDeployment) {
		log.Info("Updating Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
		if err := r.Update(ctx, foundDeployment); err != nil {
			log.Error(err, "unable to update deployment")
			return err
		}
	}

	return nil
}

// Service reconciles a k8s service object.
func Service(ctx context.Context, r client.Client, service *corev1.Service, log logr.Logger) error {
	foundService := &corev1.Service{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Service", "namespace", service.Namespace, "name", service.Name)
			if err = r.Create(ctx, service); err != nil {
				log.Error(err, "unable to create service")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "error getting service")
			return err
		}
	}
	if !justCreated && CopyServiceFields(service, foundService) {
		log.Info("Updating Service\n", "namespace", service.Namespace, "name", service.Name)
		if err := r.Update(ctx, foundService); err != nil {
			log.Error(err, "unable to update Service")
			return err
		}
	}

	return nil
}

// Namespace reconciles a Namespace object.
func Namespace(ctx context.Context, r client.Client, namespace *corev1.Namespace, log logr.Logger) error {
	foundNamespace := &corev1.Namespace{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: namespace.Name, Namespace: namespace.Namespace}, foundNamespace); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating namespace", "namespace", namespace.Name)
			if err = r.Create(ctx, namespace); err != nil {
				// IncRequestErrorCounter("error creating namespace", SEVERITY_MAJOR)
				log.Error(err, "Unable to create namespace")
				return err
			}
			err = backoff.Retry(
				func() error {
					return r.Get(ctx, types.NamespacedName{Name: namespace.Name}, foundNamespace)
				},
				backoff.WithMaxRetries(backoff.NewConstantBackOff(3*time.Second), 5))
			if err != nil {
				// IncRequestErrorCounter("error namespace create completion", SEVERITY_MAJOR)
				log.Error(err, "error namespace create completion")
				return err
				// return r.appendErrorConditionAndReturn(ctx, namespace,
				// "Owning namespace failed to create within 15 seconds")
			}
			log.Info("Created Namespace: "+foundNamespace.Name, "status", foundNamespace.Status.Phase)
			justCreated = true
		} else {
			// IncRequestErrorCounter("error reading namespace", SEVERITY_MAJOR)
			log.Error(err, "Error reading namespace")
			return err
		}
	}
	if !justCreated && CopyNamespace(namespace, foundNamespace) {
		log.Info("Updating Namespace\n", "namespace", namespace.Name)
		if err := r.Update(ctx, foundNamespace); err != nil {
			log.Error(err, "Unable to update Namespace")
			return err
		}
	}

	return nil
}

// ServiceAccount reconciles a Service Account object.
func ServiceAccount(ctx context.Context, r client.Client, serviceAccount *corev1.ServiceAccount, log logr.Logger) error {
	foundServiceAccount := &corev1.ServiceAccount{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Service Account", "namespace", serviceAccount.Namespace, "name", serviceAccount.Name)
			if err = r.Create(ctx, serviceAccount); err != nil {
				log.Error(err, "Unable to create Service Account")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Error getting Service Account")
			return err
		}
	}
	if !justCreated && CopyServiceAccount(serviceAccount, foundServiceAccount) {
		log.Info("Updating Service Account\n", "namespace", serviceAccount.Namespace, "name", serviceAccount.Name)
		if err := r.Update(ctx, foundServiceAccount); err != nil {
			log.Error(err, "Unable to update Service Account")
			return err
		}
	}

	return nil
}

// RoleBinding reconciles a Role Binding object.
func RoleBinding(ctx context.Context, r client.Client, roleBinding *rbacv1.RoleBinding, log logr.Logger) error {
	foundRoleBinding := &rbacv1.RoleBinding{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, foundRoleBinding); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating Role Binding", "namespace", roleBinding.Namespace, "name", roleBinding.Name)
			if err = r.Create(ctx, roleBinding); err != nil {
				log.Error(err, "Unable to create Role Binding")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Error getting Role Binding")
			return err
		}
	}
	if !justCreated && CopyRoleBinding(roleBinding, foundRoleBinding) {
		log.Info("Updating Role Binding\n", "namespace", roleBinding.Namespace, "name", roleBinding.Name)
		if err := r.Update(ctx, foundRoleBinding); err != nil {
			log.Error(err, "Unable to update Role Binding")
			return err
		}
	}

	return nil
}

// NetworkPolicy reconciles a NetworkPolicy object.
func NetworkPolicy(ctx context.Context, r client.Client, networkPolicy *networkv1.NetworkPolicy, log logr.Logger) error {
	foundNetworkPolicy := &networkv1.NetworkPolicy{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: networkPolicy.Name, Namespace: networkPolicy.Namespace}, foundNetworkPolicy); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating NetworkPolicy", "namespace", networkPolicy.Namespace, "name", networkPolicy.Name)
			if err = r.Create(ctx, networkPolicy); err != nil {
				log.Error(err, "Unable to create NetworkPolicy")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Error getting NetworkPolicy")
			return err
		}
	}
	if !justCreated && CopyNetworkPolicy(networkPolicy, foundNetworkPolicy) {
		log.Info("Updating NetworkPolicy\n", "namespace", networkPolicy.Namespace, "name", networkPolicy.Name)
		if err := r.Update(ctx, foundNetworkPolicy); err != nil {
			log.Error(err, "Unable to update NetworkPolicy")
			return err
		}
	}

	return nil
}

// StorageClass reconciles a StorageClass object.
func StorageClass(ctx context.Context, r client.Client, storageClass *storagev1.StorageClass, log logr.Logger) error {
	foundStorageClass := &storagev1.StorageClass{}
	justCreated := false
	if err := r.Get(ctx, types.NamespacedName{Name: storageClass.Name}, foundStorageClass); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating StorageClass", "name", storageClass.Name)
			if err = r.Create(ctx, storageClass); err != nil {
				log.Error(err, "Unable to create StorageClass")
				return err
			}
			justCreated = true
		} else {
			log.Error(err, "Error getting StorageClass")
			return err
		}
	}
	if !justCreated && CopyStorageClass(storageClass, foundStorageClass) {
		log.Info("Updating StorageClass", "name", storageClass.Name)
		if err := r.Update(ctx, foundStorageClass); err != nil {
			log.Error(err, "Unable to update StorageClass")
			return err
		}
	}

	return nil
}

// Reference: https://github.com/pwittrock/kubebuilder-workshop/blob/master/pkg/util/util.go

// CopyStatefulSetFields copies the owned fields from one StatefulSet to another
// Returns true if the fields copied from don't match to.
func CopyStatefulSetFields(from, to *appsv1.StatefulSet) bool {
	requireUpdate := false
	if !reflect.DeepEqual(to.Labels, from.Labels) {
		to.Labels = from.Labels
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Annotations, from.Annotations) {
		to.Annotations = from.Annotations
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Replicas, to.Spec.Replicas) {
		to.Spec.Replicas = from.Spec.Replicas
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Labels, from.Spec.Template.Labels) {
		to.Spec.Template.Labels = from.Spec.Template.Labels
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Annotations, from.Spec.Template.Annotations) {
		to.Spec.Template.Annotations = from.Spec.Template.Annotations
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Volumes, from.Spec.Template.Spec.Volumes) {
		to.Spec.Template.Spec.Volumes = from.Spec.Template.Spec.Volumes
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.ServiceAccountName, from.Spec.Template.Spec.ServiceAccountName) {
		to.Spec.Template.Spec.ServiceAccountName = from.Spec.Template.Spec.ServiceAccountName
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.SecurityContext, from.Spec.Template.Spec.SecurityContext) {
		to.Spec.Template.Spec.SecurityContext = from.Spec.Template.Spec.SecurityContext
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].Name, from.Spec.Template.Spec.Containers[0].Name) {
		to.Spec.Template.Spec.Containers[0].Name = from.Spec.Template.Spec.Containers[0].Name
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].Image, from.Spec.Template.Spec.Containers[0].Image) {
		to.Spec.Template.Spec.Containers[0].Image = from.Spec.Template.Spec.Containers[0].Image
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].WorkingDir, from.Spec.Template.Spec.Containers[0].WorkingDir) {
		to.Spec.Template.Spec.Containers[0].WorkingDir = from.Spec.Template.Spec.Containers[0].WorkingDir
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].Ports, from.Spec.Template.Spec.Containers[0].Ports) {
		to.Spec.Template.Spec.Containers[0].Ports = from.Spec.Template.Spec.Containers[0].Ports
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].Env, from.Spec.Template.Spec.Containers[0].Env) {
		to.Spec.Template.Spec.Containers[0].Env = from.Spec.Template.Spec.Containers[0].Env
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].Resources, from.Spec.Template.Spec.Containers[0].Resources) {
		to.Spec.Template.Spec.Containers[0].Resources = from.Spec.Template.Spec.Containers[0].Resources
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec.Containers[0].VolumeMounts, from.Spec.Template.Spec.Containers[0].VolumeMounts) {
		to.Spec.Template.Spec.Containers[0].VolumeMounts = from.Spec.Template.Spec.Containers[0].VolumeMounts
		requireUpdate = true
	}

	return requireUpdate
}

func CopyDeploymentSetFields(from, to *appsv1.Deployment) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	if from.Spec.Replicas != to.Spec.Replicas {
		to.Spec.Replicas = from.Spec.Replicas
		requireUpdate = true
	}

	if !reflect.DeepEqual(to.Spec.Template.Spec, from.Spec.Template.Spec) {
		requireUpdate = true
	}
	to.Spec.Template.Spec = from.Spec.Template.Spec

	return requireUpdate
}

// CopyServiceFields copies the owned fields from one Service to another
func CopyServiceFields(from, to *corev1.Service) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	// Don't copy the entire Spec, because we can't overwrite the clusterIp field

	if !reflect.DeepEqual(to.Spec.Selector, from.Spec.Selector) {
		requireUpdate = true
	}
	to.Spec.Selector = from.Spec.Selector

	if !reflect.DeepEqual(to.Spec.Ports, from.Spec.Ports) {
		requireUpdate = true
	}
	to.Spec.Ports = from.Spec.Ports

	return requireUpdate
}

// CopyServiceAccount copies the owned fields from one Service Account to another
func CopyServiceAccount(from, to *corev1.ServiceAccount) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	// Don't copy the entire Spec, because we this will lead to unnecessary reconciles
	if !reflect.DeepEqual(to.ImagePullSecrets, from.ImagePullSecrets) {
		requireUpdate = true
	}
	to.ImagePullSecrets = from.ImagePullSecrets

	if !reflect.DeepEqual(to.AutomountServiceAccountToken, from.AutomountServiceAccountToken) {
		requireUpdate = true
	}
	to.AutomountServiceAccountToken = from.AutomountServiceAccountToken

	return requireUpdate
}

// CopyRoleBinding copies the owned fields from one Role Binding to another
func CopyRoleBinding(from, to *rbacv1.RoleBinding) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	// Don't copy the entire Spec, because we this will lead to unnecessary reconciles
	if !reflect.DeepEqual(to.RoleRef, from.RoleRef) {
		requireUpdate = true
	}
	to.RoleRef = from.RoleRef

	if !reflect.DeepEqual(to.Subjects, from.Subjects) {
		requireUpdate = true
	}
	to.Subjects = from.Subjects

	return requireUpdate
}

// CopyNetworkPolicy copies the owned fields from one NetworkPolicy to another
func CopyNetworkPolicy(from, to *networkv1.NetworkPolicy) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	if !reflect.DeepEqual(to.Spec, from.Spec) {
		requireUpdate = true
	}
	to.Spec = from.Spec

	return requireUpdate
}

// CopyNamespace copies the owned fields from one Namespace to another
func CopyNamespace(from, to *corev1.Namespace) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	return requireUpdate
}

// CopyStorageClass copies the owned fields from one StorageClass to another
func CopyStorageClass(from, to *storagev1.StorageClass) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Labels) == 0 && len(from.Labels) != 0 {
		requireUpdate = true
	}
	to.Labels = from.Labels

	for k, v := range to.Annotations {
		if from.Annotations[k] != v {
			requireUpdate = true
		}
	}
	if len(to.Annotations) == 0 && len(from.Annotations) != 0 {
		requireUpdate = true
	}
	to.Annotations = from.Annotations

	return requireUpdate
}
