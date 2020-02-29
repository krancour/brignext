package brignext

type ProjectKubernetesConfig struct {
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty"`
}
