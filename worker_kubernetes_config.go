package brignext

type WorkerKubernetesConfig struct {
	ImagePullSecrets []string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}
