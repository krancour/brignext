package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
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
	Container            ContainerSpec          `json:"container" bson:"container"`
	WorkspaceSize        string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git                  WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes           WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	Jobs                 JobsSpec               `json:"jobs" bson:"jobs"`
	LogLevel             LogLevel               `json:"logLevel" bson:"logLevel"`
	ConfigFilesDirectory string                 `json:"configFilesDirectory" bson:"configFilesDirectory"`
	DefaultConfigFiles   map[string]string      `json:"defaultConfigFiles" bson:"defaultConfigFiles"`
}

type ContainerSpec struct {
	Image           string `json:"image" bson:"image"`
	ImagePullPolicy string `json:"imagePullPolicy" bson:"imagePullPolicy"`
	// Command specifies the command to be executed by the OCI container. This
	// can be used to optionally override the default command specified by the OCI
	// image itself.
	Command []string `json:"command" bson:"command"`
	// Arguments specifies arguments to the command executed by the OCI container.
	Arguments   []string          `json:"arguments" bson:"arguments"`
	Environment map[string]string `json:"environment" bson:"environment"`
}

type WorkerGitConfig struct {
	CloneURL       string `json:"cloneURL" bson:"cloneURL"`
	Commit         string `json:"commit" bson:"commit"`
	Ref            string `json:"ref" bson:"ref"`
	InitSubmodules bool   `json:"initSubmodules" bson:"initSubmodules"`
}

type WorkerKubernetesConfig struct {
	ImagePullSecrets []string `json:"imagePullSecrets" bson:"imagePullSecrets"`
}

type WorkerStatus struct {
	Started *time.Time  `json:"started" bson:"started"`
	Ended   *time.Time  `json:"ended" bson:"ended"`
	Phase   WorkerPhase `json:"phase" bson:"phase"`
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
