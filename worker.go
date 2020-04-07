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
	Container     ContainerConfig        `json:"container" bson:"container"`
	WorkspaceSize string                 `json:"workspaceSize" bson:"workspaceSize"`
	Git           WorkerGitConfig        `json:"git" bson:"git"`
	Kubernetes    WorkerKubernetesConfig `json:"kubernetes" bson:"kubernetes"`
	JobsConfig    JobsConfig             `json:"jobsConfig" bson:"jobsConfig"`
	LogLevel      LogLevel               `json:"logLevel" bson:"logLevel"`
	Started       *time.Time             `json:"started" bson:"started"`
	Ended         *time.Time             `json:"ended" bson:"ended"`
	Status        WorkerStatus           `json:"status" bson:"status"`
}
