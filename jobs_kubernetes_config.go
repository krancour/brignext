package brignext

type JobsKubernetesConfig struct {
	ImagePullSecrets string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}
