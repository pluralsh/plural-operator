package redeployment

type Resource string

const (
	ResourceConfigMap = "configmap"
	ResourceSecret    = "secret"
)

// Annotations used by the redeployment operator
const (
	RedeployLabel = "platform.plural.sh/redeploy"
	ShaAnnotation = "platform.plural.sh/sha"
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
	IsControlled() bool

	// HasAnnotation checks if controlled Resource contains SHA annotation.
	HasAnnotation() bool

	// UpdateAnnotation updates the SHA annotation with the latest SHA
	// calculated based on the Resource data. It is used to determine
	// if a Resource has been updated.
	UpdateAnnotation() error

	// ShouldDeletePods checks if existing SHA annotation is in sync with the actual SHA
	// calculated based on the latest Resource data.
	ShouldDeletePods() bool

	// DeletePods deletes all matching pods
	DeletePods() error

	// getSHA calculates the SHA of the Resource data.
	getSHA() string
}
