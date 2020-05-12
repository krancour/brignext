package brignext

import "time"

type WorkerPhase string

const (
	// WorkerPhasePending represents the state wherein a worker is awaiting
	// execution.
	WorkerPhasePending WorkerPhase = "PENDING"
	// WorkerPhaseRunning represents the state wherein a worker is currently
	// being executed.
	WorkerPhaseRunning WorkerPhase = "RUNNING"
	// WorkerPhaseCanceled represents the state wherein a pending worker was
	// canceled prior to execution.
	WorkerPhaseCanceled WorkerPhase = "CANCELED"
	// WorkerPhaseAborted represents the state wherein a worker was forcefully
	// stopped during execution.
	WorkerPhaseAborted WorkerPhase = "ABORTED"
	// WorkerPhaseSucceeded represents the state where a worker has run to
	// completion without error.
	WorkerPhaseSucceeded WorkerPhase = "SUCCEEDED"
	// WorkerPhaseFailed represents the state wherein a worker has run to
	// completion but experienced errors.
	WorkerPhaseFailed WorkerPhase = "FAILED"
	// WorkerPhaseTimedOut represents the state wherein a worker has has not
	// completed within a designated timeframe.
	WorkerPhaseTimedOut WorkerPhase = "TIMED_OUT"
	// WorkerPhaseUnknown represents the state wherein a worker's state is
	// unknown.
	WorkerPhaseUnknown WorkerPhase = "UNKNOWN"
)

// nolint: lll
type Worker struct {
	Container            ContainerConfig        `json:"container" bson:"container"`
	WorkspaceSize        string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git                  WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes           WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	JobsConfig           JobsConfig             `json:"jobsConfig" bson:"jobsConfig"`
	LogLevel             LogLevel               `json:"logLevel" bson:"logLevel"`
	ConfigFilesDirectory string                 `json:"configFilesDirectory" bson:"configFilesDirectory"`
	DefaultConfigFiles   map[string]string      `json:"defaultConfigFiles" bson:"defaultConfigFiles"`
	Jobs                 map[string]Job         `json:"jobs" bson:"jobs"`
	Status               WorkerStatus           `json:"status" bson:"status"`
}

type WorkerStatus struct {
	Started *time.Time  `json:"started" bson:"started"`
	Ended   *time.Time  `json:"ended" bson:"ended"`
	Phase   WorkerPhase `json:"phase" bson:"phase"`
}
