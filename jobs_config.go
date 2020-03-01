package brignext

type JobsConfig struct {
	AllowPrivileged bool                  `json:"allowPrivileged" bson:"allowPrivileged"`
	AllowHostMounts bool                  `json:"allowHostMounts" bson:"allowHostMounts"`
	Kubernetes      *JobsKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes"`
}
