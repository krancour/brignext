package brignext

type ContainerConfig struct {
	Image           string `json:"image" bson:"image"`
	ImagePullPolicy string `json:"imagePullPolicy" bson:"imagePullPolicy"`
	Command         string `json:"command" bson:"command"`
	// TODO: Should be able to add non-sensitive environment variables here
}
