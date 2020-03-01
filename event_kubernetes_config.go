package brignext

type EventKubernetesConfig struct {
	Namespace string `json:"namespace,omitempty" bson:"namespace"`
}
