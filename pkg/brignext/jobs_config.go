package brignext

type JobsConfig struct {
	AllowPrivileged bool `json:"allowPrivileged,omitempty" bson:"allowPrivileged,omitempty"`
	AllowHostMounts bool `json:"allowHostMounts,omitempty" bson:"allowHostMounts,omitempty"`
}
