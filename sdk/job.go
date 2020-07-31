package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// JobPhase represents where a Job is within its lifecycle.
type JobPhase string

const (
	// JobPhaseAborted represents the state wherein a job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseFailed represents the state wherein a job has run to
	// completion but experienced errors.
	JobPhaseFailed JobPhase = "FAILED"
	// JobPhaseRunning represents the state wherein a job is currently
	// being executed.
	JobPhaseRunning JobPhase = "RUNNING"
	// JobPhaseSucceeded represents the state where a job has run to
	// completion without error.
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
	// JobPhaseTimedOut represents the state wherein a job has has not completed
	// within a designated timeframe.
	JobPhaseTimedOut JobPhase = "TIMED_OUT"
	// JobPhaseUnknown represents the state wherein a job's state is unknown. Note
	// that this is possible if and only if the underlying Job execution substrate
	// (Kubernetes), for some unanticipated, reason does not know the Job's
	// (Pod's) state.
	JobPhaseUnknown WorkerPhase = "UNKNOWN"
)

// JobsSpec represents configuration and policies for any Jobs spawned by
// a Worker.
type JobsSpec struct {
	AllowPrivileged        bool                 `json:"allowPrivileged"`
	AllowDockerSocketMount bool                 `json:"allowDockerSocketMount"`
	Kubernetes             JobsKubernetesConfig `json:"kubernetes"`
}

// JobsKubernetesConfig represents Kubernetes-specific Jobs configuration.
type JobsKubernetesConfig struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which the Jobs' containers are based. The
	// image pull secrets in question must be created out-of-band by a
	// sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// JobStatus represents the status of a Job.
type JobStatus struct {
	// Started indicates the time the Job began execution.
	Started *time.Time `json:"started"`
	// Started indicates the time the Job concluded execution. It will be nil
	// for a Job that is not done executing.
	Ended *time.Time `json:"ended"`
	// Phase indicates where the Job is in its lifecycle.
	Phase JobPhase `json:"phase"`
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
