package brignext

type JobStatus string

const (
	// JobStatusPending represents the state wherein a jon is awaiting
	// execution.
	JobStatusPending JobStatus = "PENDING"
	// JobStatusRunning represents the state wherein a job is currently
	// being executed.
	JobStatusRunning JobStatus = "RUNNING"
	// JobStatusAborted represents the state wherein a job was forcefully
	// stopped during execution.
	JobStatusAborted JobStatus = "ABORTED"
	// JobStatusSucceeded represents the state where a job has run to
	// completion without error.
	JobStatusSucceeded JobStatus = "SUCCEEDED"
	// JobStatusFailed represents the state wherein a job has run to
	// completion but experienced errors.
	JobStatusFailed JobStatus = "FAILED"
)

// // Job is a single job that is executed by the worker that processes and event.
// type Job struct {
// 	ID           string    `json:"id,omitempty" bson:"_id,"`
// 	Name         string    `json:"name,omitempty" bson:"name,"`
// 	EventID      string    `json:"eventID,omitempty" bson:"eventID,"`
// 	Image        string    `json:"image,omitempty" bson:"image,"`
// 	CreationTime time.Time `json:"creationTime,omitempty" bson:"creationTime,"`
// 	StartTime    time.Time `json:"startTime,omitempty" bson:"startTime,"`
// 	EndTime      time.Time `json:"endTime,omitempty" bson:"endTime,"`
// 	ExitCode     int32     `json:"exitCode,omitempty" bson:"exitCode,"`
// 	Status       JobStatus `json:"status,omitempty" bson:"status,"`
// }
