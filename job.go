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
	// JobPhaseTimedOut represents the state wherein a job has has not completed
	// within a designated timeframe.
	JobPhaseTimedOut JobPhase = "TIMED_OUT"
	// JobPhaseUnknown represents the state wherein a job's state is unknown.
	JobPhaseUnknown JobPhase = "UNKNOWN"
)

// nolint: lll
type JobsSpec struct {
	AllowPrivileged        bool                 `json:"allowPrivileged" bson:"allowPrivileged"`
	AllowDockerSocketMount bool                 `json:"allowDockerSocketMount" bson:"allowDockerSocketMount"`
	Kubernetes             JobsKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
}

type JobsKubernetesConfig struct {
	ImagePullSecrets []string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}

type JobStatus struct {
	Started *time.Time `json:"started" bson:"started"`
	Ended   *time.Time `json:"ended" bson:"ended"`
	Phase   JobPhase   `json:"phase" bson:"phase"`
}
