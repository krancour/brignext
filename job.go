package brignext

import "time"

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
	// JobStatusUnknown represents the state wherein a job's status is unknown.
	JobStatusUnknown JobStatus = "UNKNOWN"
)

// Job is a single job that is executed by a worker that processes an event.
type Job struct {
	Started *time.Time `json:"started" bson:"started"`
	Ended   *time.Time `json:"ended" bson:"ended"`
	Status  JobStatus  `json:"status,omitempty" bson:"status,"`
}
