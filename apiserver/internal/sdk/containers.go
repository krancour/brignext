package sdk

// ImagePullPolicy represents a policy for whether container hosts already
// having a certain OCI image should attempt to re-pull that image prior to
// launching a new container based on that image.
type ImagePullPolicy string

const (
	// ImagePullPolicyIfNotPresent represents a policy wherein container hosts
	// only attempt to pull an OCI image if that image does not already exist on
	// the host.
	ImagePullPolicyIfNotPresent ImagePullPolicy = "IfNotPresent"
	// ImagePullPolicyAlways represents a policy wherein container hosts will
	// always attempt to re-pull an OCI image before launching a container based
	// on that image.
	ImagePullPolicyAlways ImagePullPolicy = "Always"
)

// ContainerSpec represents the technical details of an OCI container.
type ContainerSpec struct {
	// Image specified the OCI image on which the container should be based.
	Image string `json:"image,omitempty" bson:"image,omitempty"`
	// ImagePullPolicy specifies whether a container host already having the
	// specified OCI image should attempt to re-pull that image prior to launching
	// a new container.
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy,omitempty" bson:"imagePullPolicy,omitempty"` // nolint: lll
	// Command specifies the command to be executed by the OCI container. This
	// can be used to optionally override the default command specified by the OCI
	// image itself.
	Command []string `json:"command,omitempty" bson:"command,omitempty"`
	// Arguments specifies arguments to the command executed by the OCI container.
	Arguments []string `json:"arguments,omitempty" bson:"arguments,omitempty"`
	// Environment is a map of key/value pairs that specify environment variables
	// to be set within the OCI container.
	Environment map[string]string `json:"environment,omitempty" bson:"environment,omitempty"` // nolint: lll
}

// nolint: lll
type JobContainerSpec struct {
	ContainerSpec      `json:",inline" bson:",inline"`
	UseWorkspace       bool   `json:"useWorkspace" bson:"useWorkspace"`
	WorkspaceMountPath string `json:"workspaceMountPath,omitempty" bson:"workspaceMountPath,omitempty"`
	UseSource          bool   `json:"useSource" bson:"useSource"`
	SourceMountPath    string `json:"sourceMountPath,omitempty" bson:"sourceMountPath,omitempty"`
	Privileged         bool   `json:"privileged" bson:"privileged"`
	DockerSocketMount  bool   `json:"dockerSocketMount" bson:"dockerSocketMount"`
}
