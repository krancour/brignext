package brignext

// nolint: lll
type WorkerKubernetesConfig struct {
	WorkspaceStorageClass string   `json:"workspaceStorageClass" bson:"workspaceStorageClass"`
	ImagePullSecrets      []string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}
