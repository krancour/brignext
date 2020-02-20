package brignext

type ContainerConfig struct {
	Image   *Image `json:"image,omitempty" bson:"image,omitempty"`
	Command string `json:"command,omitempty" bson:"command,omitempty"`
}
