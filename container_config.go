package brignext

// nolint: lll
type ContainerConfig struct {
	Image           string            `json:"image" bson:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy" bson:"imagePullPolicy"`
	Command         string            `json:"command" bson:"command"`
	Environment     map[string]string `json:"environment" bson:"environment"`
}
