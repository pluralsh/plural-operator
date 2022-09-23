package redeployment

// Annotations used by the oauthinjector_webhook operator
const (
	RedeployLabel = "security.plural.sh/inject-oauth-sidecar"
)

// Service is a redeployment operator service that simplifies the process of
// finding the application that uses a Resource and triggering a rollout restart
// of the application in order to be able to always use the latest config values
// without having to manually restart the application after every change.
type Service interface {
	// IsControlled checks if secret should be controlled by the
	// redeployment operator.
	IsControlled() bool

	// DeletePods deletes all matching pods
	DeletePods() error
}
