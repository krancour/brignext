package brignext

// nolint: lll
type JobsConfig struct {
	AllowPrivileged        bool                 `json:"allowPrivileged" bson:"allowPrivileged"`
	AllowDockerSocketMount bool                 `json:"allowDockerSocketMount" bson:"allowDockerSocketMount"`
	Kubernetes             JobsKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
}
