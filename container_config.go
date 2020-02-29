package brignext

// nolint: lll
type ContainerConfig struct {
	Image           string `json:"image,omitempty" bson:"image,omitempty"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty" bson:"imagePullPolicy,omitempty"`
	Command         string `json:"command,omitempty" bson:"command,omitempty"`
}
