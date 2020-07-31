package sdk

// Labels is a map of key/value pairs utilized by Events in describing
// themselves and by EventSubscriptions in describing Events of interest.
type Labels map[string]string

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
