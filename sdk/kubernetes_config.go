package sdk

// KubernetesConfig represents Kubernetes-specific configuration. This is used
// primarily at the Project level, but is also denormalized onto Events so that
// Event handling doesn't required a Project lookup to obtain
// Kubernetes-specific configuration.
type KubernetesConfig struct {
	Namespace string `json:"namespace"`
}
