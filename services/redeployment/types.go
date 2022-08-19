package redeployment

import (
	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
)

type Resource string

const (
	ResourceConfigMap = "configmap"
	ResourceSecret    = "secret"
)

// Annotations used by the redeployment operator
const (
	shaAnnotation     = "platform.plural.sh/sha"
	restartAnnotation = "kubectl.kubernetes.io/restartedAt"
)

// Service is a redeployment operator service that simplifies the process of
// finding the application that uses a Resource and triggering a rollout restart
// of the application in order to be able to always use the latest config values
// without having to manually restart the application after every change.
//
// Currently supported resources:
//   - ResourceConfigMap
//   - ResourceSecret
type Service interface {
	// IsControlled checks if the v1alpha1.Redeployment Resource should be controlled by the
	// redeployment operator.
	IsControlled() (bool, error)

	// HasAnnotation checks if controlled Resource contains SHA annotation.
	HasAnnotation() bool

	// UpdateAnnotation updates the SHA annotation with the latest SHA
	// calculated based on the Resource data. It is used to determine
	// if a Resource has been updated.
	UpdateAnnotation() error

	// ShouldRestart checks if existing SHA annotation is in sync with the actual SHA
	// calculated based on the latest Resource data.
	ShouldRestart() bool

	// RolloutRestart triggers a rollout restart with a zero downtime for the application
	// that controls the v1alpha1.Redeployment Resource based on the v1alpha1.WorkflowType.
	RolloutRestart() error

	// getSHA calculates the SHA of the Resource data.
	getSHA() string

	// isControlled TODO
	isControlled(redeployment *v1alpha1.Redeployment) (bool, error)
}

// Workflow TODO
type Workflow interface {
	// IsUsed checks if provided Resource is used by the v1alpha1.WorkflowType
	// based on the volumes attached to that v1alpha1.WorkflowType resource.
	IsUsed(resource Resource, namespace string, name string) bool

	// RolloutRestart TODO
	RolloutRestart(redeployment *v1alpha1.Redeployment) error
}
