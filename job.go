package brignext

import "time"

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
	*TypeMeta `json:",inline,omitempty" bson:"-"`
	Started   *time.Time `json:"started" bson:"started"`
	Ended     *time.Time `json:"ended" bson:"ended"`
	Phase     JobPhase   `json:"phase" bson:"phase"`
}

func NewJobStatus() JobStatus {
	return JobStatus{
		TypeMeta: &TypeMeta{
			APIVersion: APIVersion,
			Kind:       "JobStatus",
		},
	}
}
