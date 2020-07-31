package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

type WorkerPhase string

const (
	// WorkerPhaseAborted represents the state wherein a worker was forcefully
	// stopped during execution.
	WorkerPhaseAborted WorkerPhase = "ABORTED"
	// WorkerPhaseCanceled represents the state wherein a pending worker was
	// canceled prior to execution.
	WorkerPhaseCanceled WorkerPhase = "CANCELED"
	// WorkerPhaseFailed represents the state wherein a worker has run to
	// completion but experienced errors.
	WorkerPhaseFailed WorkerPhase = "FAILED"
	// WorkerPhasePending represents the state wherein a worker is awaiting
	// execution.
	WorkerPhasePending WorkerPhase = "PENDING"
	// WorkerPhaseRunning represents the state wherein a worker is currently
	// being executed.
	WorkerPhaseRunning WorkerPhase = "RUNNING"
	// WorkerPhaseSucceeded represents the state where a worker has run to
	// completion without error.
	WorkerPhaseSucceeded WorkerPhase = "SUCCEEDED"
	// WorkerPhaseTimedOut represents the state wherein a worker has has not
	// completed within a designated timeframe.
	WorkerPhaseTimedOut WorkerPhase = "TIMED_OUT"
	// WorkerPhaseUnknown represents the state wherein a worker's state is
	// unknown.
	WorkerPhaseUnknown WorkerPhase = "UNKNOWN"
)

func WorkerPhasesAll() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhaseAborted,
		WorkerPhaseCanceled,
		WorkerPhaseFailed,
		WorkerPhasePending,
		WorkerPhaseRunning,
		WorkerPhaseSucceeded,
		WorkerPhaseTimedOut,
		WorkerPhaseUnknown,
	}
}

func WorkerPhasesTerminal() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhaseAborted,
		WorkerPhaseCanceled,
		WorkerPhaseFailed,
		WorkerPhaseSucceeded,
		WorkerPhaseTimedOut,
	}
}

func WorkerPhasesNonTerminal() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhasePending,
		WorkerPhaseRunning,
		WorkerPhaseUnknown,
	}
}

// nolint: lll
type WorkerSpec struct {
	Container            ContainerSpec          `json:"container"`
	WorkspaceSize        string                 `json:"workspaceSize"`
	Git                  WorkerGitConfig        `json:"git"`
	Kubernetes           WorkerKubernetesConfig `json:"kubernetes"`
	Jobs                 JobsSpec               `json:"jobs"`
	LogLevel             LogLevel               `json:"logLevel"`
	ConfigFilesDirectory string                 `json:"configFilesDirectory"`
	DefaultConfigFiles   map[string]string      `json:"defaultConfigFiles"`
}

type ContainerSpec struct {
	Image           string            `json:"image"`
	ImagePullPolicy string            `json:"imagePullPolicy"` // nolint: lll
	Command         string            `json:"command"`
	Environment     map[string]string `json:"environment"`
}

type WorkerGitConfig struct {
	CloneURL       string `json:"cloneURL"`
	Commit         string `json:"commit"`
	Ref            string `json:"ref"`
	InitSubmodules bool   `json:"initSubmodules"`
}

type WorkerKubernetesConfig struct {
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

type WorkerStatus struct {
	Started *time.Time  `json:"started"`
	Ended   *time.Time  `json:"ended"`
	Phase   WorkerPhase `json:"phase"`
}

func (w WorkerStatus) MarshalJSON() ([]byte, error) {
	type Alias WorkerStatus
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "WorkerStatus",
			},
			Alias: (Alias)(w),
		},
	)
}
