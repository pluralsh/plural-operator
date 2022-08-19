package controllers

const (
	managedLabel                = "app.kubernetes.io/managed-by"
	ignoreAnnotation            = "platform.plural.sh/ignore"
	shaAnnotation               = "platform.plural.sh/sha"
	deploymentRestartAnnotation = "kubectl.kubernetes.io/restartedAt"
)
