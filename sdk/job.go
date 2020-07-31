package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

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
	AllowPrivileged        bool                 `json:"allowPrivileged"`
	AllowDockerSocketMount bool                 `json:"allowDockerSocketMount"`
	Kubernetes             JobsKubernetesConfig `json:"kubernetes"`
}

type JobsKubernetesConfig struct {
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

type JobStatus struct {
	Started *time.Time `json:"started"`
	Ended   *time.Time `json:"ended"`
	Phase   JobPhase   `json:"phase"`
}

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
