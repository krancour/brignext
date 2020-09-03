package api

// TODO: Document this
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string
}

type LogStreamOptions struct {
	Follow bool `json:"follow"`
}
