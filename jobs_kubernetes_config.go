package brignext

type JobsKubernetesConfig struct {
	CacheStorageClass string `json:"cacheStorageClass" bson:"cacheStorageClass"`
	AllowSecretKeyRef bool   `json:"allowSecretKeyRef" bson:"allowSecretKeyRef"`
	// TODO: This can be done more elegantly than making the user name a secret
	// that's been created out of band. i.e. We should be able to manage this
	// for them if we create the right APIs. This is fine for prototyping
	// purposes.
	ImagePullSecrets string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}
