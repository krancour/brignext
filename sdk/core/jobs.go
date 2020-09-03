package core

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// JobPhase represents where a Job is within its lifecycle.
type JobPhase string

const (
	// JobPhaseAborted represents the state wherein a Job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseFailed represents the state wherein a Job has run to
	// completion but experienced errors.
	JobPhaseFailed JobPhase = "FAILED"
	// JobPhasePending represents the state wherein a Job is awaiting
	// execution.
	JobPhasePending JobPhase = "PENDING"
	// JobPhaseRunning represents the state wherein a Job is currently
	// being executed.
	JobPhaseRunning JobPhase = "RUNNING"
	// JobPhaseSucceeded represents the state where a Job has run to
	// completion without error.
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
	// JobPhaseTimedOut represents the state wherein a Job has has not completed
	// within a designated timeframe.
	JobPhaseTimedOut JobPhase = "TIMED_OUT"
	// JobPhaseUnknown represents the state wherein a Job's state is unknown. Note
	// that this is possible if and only if the underlying Job execution substrate
	// (Kubernetes), for some unanticipated, reason does not know the Job's
	// (Pod's) state.
	JobPhaseUnknown WorkerPhase = "UNKNOWN"
)

// Job represents a component spawned by a Worker to complete a single task
// during the handling of an Event.
type Job struct {
	// Spec is the technical blueprint for the Job.
	Spec JobSpec `json:"spec"`
	// Status contains details of the Job's current state.
	Status JobStatus `json:"status"`
}

// JobSpec is the technical blueprint for a Job.
type JobSpec struct {
	// PrimaryContainer specifies the details of an OCI container that forms the
	// cornerstone of the Job. Job success or failure is tied to completion and
	// exit code of this container.
	PrimaryContainer JobContainerSpec `json:"primaryContainer"`
	// SidecarContainers specifies the details of supplemental, "sidecar"
	// containers. Their completion and exit code do not directly impact Job
	// status. BrigNext does not understand dependencies between a Job's multiple
	// containers and cannot enforce any specific startup or shutdown order. When
	// such dependencies exist (for instance, a primary container than cannot
	// proceed with a suite of tests until a database is launched and READY in a
	// sidecar container), then logic within those containers must account for
	// these constraints.
	SidecarContainers map[string]JobContainerSpec `json:"sidecarContainers,omitempty"` // nolint: lll
	// TODO: Document this
	TimeoutSeconds int64 `json:"timeoutSeconds,omitempty"`
	// TODO: Document this
	Host *JobHost `json:"host,omitempty"`
}

// TODO: Document this
type JobHost struct {
	// TODO: Document this
	OS string `json:"os,omitempty"`
	// TODO: Document this
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// JobStatus represents the status of a Job.
type JobStatus struct {
	// Started indicates the time the Job began execution.
	Started *time.Time `json:"started,omitempty"`
	// Ended indicates the time the Job concluded execution. It will be nil
	// for a Job that is not done executing.
	Ended *time.Time `json:"ended,omitempty"`
	// Phase indicates where the Job is in its lifecycle.
	Phase JobPhase `json:"phase,omitempty"`
}

// MarshalJSON amends JobStatus instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (j JobStatus) MarshalJSON() ([]byte, error) {
	type Alias JobStatus
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "JobStatus",
			},
			Alias: (Alias)(j),
		},
	)
}
