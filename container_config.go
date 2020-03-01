package brignext

// nolint: lll
type ContainerConfig struct {
	Image           string `json:"image,omitempty" bson:"image"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty" bson:"imagePullPolicy"`
	Command         string `json:"command,omitempty" bson:"command"`
}
