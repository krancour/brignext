package brignext

// nolint: lll
type KubernetesConfig struct {
	WorkerStorageClass string `json:"workerStorageClass,omitempty" bson:"workerStorageClass,omitempty"`
	WorkerStorageSize  string `json:"workerStorageSize,omitempty" bson:"workerStorageSize,omitempty"`
	CacheStorageClass  string `json:"cacheStorageClass,omitempty" bson:"cacheStorageClass,omitempty"`
	// AllowSecretKeyRef bool   `json:"allowSecretKeyRef,omitempty" bson:"allowSecretKeyRef,omitempty"`
}
