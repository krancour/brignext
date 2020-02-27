package brignext

// nolint: lll
type KubernetesConfig struct {
	WorkspaceStorageClass string `json:"workspaceStorageClass,omitempty" bson:"workspaceStorageClass,omitempty"`
	ServiceAccount        string `json:"serviceAccount,omitempty" bson:"serviceAccount,omitempty"`
}
