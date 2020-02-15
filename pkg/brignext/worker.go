package brignext

import "time"

type WorkerStatus string

const (
	// WorkerStatusPending represents the state wherein a worker is awaiting
	// execution.
	WorkerStatusPending WorkerStatus = "PENDING"
	// WorkerStatusCanceled represents the state wherein a pending worker was
	// canceled prior to execution.
	WorkerStatusCanceled WorkerStatus = "CANCELED"
	// WorkerStatusRunning represents the state wherein a worker is currently
	// being executed.
	WorkerStatusRunning WorkerStatus = "RUNNING"
	// WorkerStatusAborted represents the state wherein a worker was forcefully
	// stopped during execution.
	WorkerStatusAborted WorkerStatus = "ABORTED"
	// WorkerStatusSucceeded represents the state where a worker has run to
	// completion without error.
	WorkerStatusSucceeded WorkerStatus = "SUCCEEDED"
	// WorkerStatusFailed represents the state wherein a worker has run to
	// completion but experienced errors.
	WorkerStatusFailed WorkerStatus = "FAILED"
)

// nolint: lll
type Worker struct {
	InitContainer *ContainerConfig `json:"initContainer,omitempty" bson:"initContainer,omitempty"`
	Container     *ContainerConfig `json:"container,omitempty" bson:"container,omitempty"`
	Jobs          *JobsConfig      `json:"jobs,omitempty" bson:"jobs,omitempty"`
	Started       *time.Time       `json:"started,omitempty" bson:"started,omitempty"`
	Ended         *time.Time       `json:"ended,omitempty" bson:"ended,omitempty"`
	Status        WorkerStatus     `json:"status,omitempty" bson:"status,omitempty"`
	// ExitCode  *int32       `json:"exitCode,omitempty" bson:"exitCode,omitempty"`
}
