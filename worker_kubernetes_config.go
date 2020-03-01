package brignext

// nolint: lll
type WorkerKubernetesConfig struct {
	WorkspaceStorageClass string `json:"workspaceStorageClass,omitempty" bson:"workspaceStorageClass"`
	ServiceAccount        string `json:"serviceAccount,omitempty" bson:"serviceAccount"`
	// TODO: This can be done more elegantly than making the user name a secret
	// that's been created out of band. i.e. We should be able to manage this
	// for them if we create the right APIs. This is fine for prototyping
	// purposes.
	ImagePullSecrets string `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets"`
}
