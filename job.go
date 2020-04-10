package brignext

import "time"

type JobPhase string

const (
	// JobPhaseRunning represents the state wherein a job is currently
	// being executed.
	JobPhaseRunning JobPhase = "RUNNING"
	// JobPhaseAborted represents the state wherein a job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseSucceeded represents the state where a job has run to
	// completion without error.
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
	// JobPhaseFailed represents the state wherein a job has run to
	// completion but experienced errors.
	JobPhaseFailed JobPhase = "FAILED"
	// JobPhaseUnknown represents the state wherein a job's state is unknown.
	JobPhaseUnknown JobPhase = "UNKNOWN"
)

// Job is a single job that is executed by a worker that processes an event.
type Job struct {
	Status JobStatus `json:"status" bson:"status"`
}

type JobStatus struct {
	Started *time.Time `json:"started" bson:"started"`
	Ended   *time.Time `json:"ended" bson:"ended"`
	Phase   JobPhase   `json:"phase" bson:"phase"`
}
