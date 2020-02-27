package brignext

type JobsKubernetesConfig struct {
	CacheStorageClass string `json:"cacheStorageClass,omitempty" bson:"cacheStorageClass,omitempty"`
	AllowSecretKeyRef bool   `json:"allowSecretKeyRef,omitempty" bson:"allowSecretKeyRef,omitempty"`
	ServiceAccount    string `json:"serviceAccount,omitempty" bson:"serviceAccount,omitempty"`
}
