package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// LogLevel represents the desired granularity of Worker log output.
type LogLevel string

const (
	// LogLevelDebug represents DEBUG level granularity in Worker log output.
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelInfo represents INFO level granularity in Worker log output.
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn represents WARN level granularity in Worker log output.
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError represents ERROR level granularity in Worker log output.
	LogLevelError LogLevel = "ERROR"
)

// ImagePullPolicy represents a policy for whether container hosts already
// having a certain OCI image should attempt to re-pull that image prior to
// launching a new container based on that image.
type ImagePullPolicy string

const (
	// ImagePullPolicyIfNotPresent represents a policy wherein container hosts
	// only attempt to pull an OCI image if that image does not already exist on
	// the host.
	ImagePullPolicyIfNotPresent ImagePullPolicy = "IfNotPresent"
	// ImagePullPolicyAlways represents a policy wherein container hosts will
	// always attempt to re-pull an OCI image before launching a container based
	// on that image.
	ImagePullPolicyAlways ImagePullPolicy = "Always"
)

// WorkerPhase represents where a Worker is within its lifecycle.
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
	// unknown. Note that this is possible if and only if the underlying Worker
	// execution substrate (Kubernetes), for some unanticipated, reason does not
	// know the Worker's (Pod's) state.
	WorkerPhaseUnknown WorkerPhase = "UNKNOWN"
)

// WorkerPhasesAll returns a slice of WorkerPhases containing ALL possible
// phases. Note that instead of utilizing a package-level slice, this a function
// returns ad-hoc copies of the slice in order to preculde the possibility of
// this important collection being modified at runtime by a client.
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

// WorkerPhasesTerminal returns a slice of WorkerPhases containing ALL phases
// that are considered terminal. Note that instead of utilizing a package-level
// slice, this a function returns ad-hoc copies of the slice in order to
// preculde the possibility of this important collection being modified at
// runtime by a client.
func WorkerPhasesTerminal() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhaseAborted,
		WorkerPhaseCanceled,
		WorkerPhaseFailed,
		WorkerPhaseSucceeded,
		WorkerPhaseTimedOut,
	}
}

// WorkerPhasesNonTerminal returns a slice of WorkerPhases containing ALL phases
// that are considered non-terminal. Note that instead of utilizing a
// package-level slice, this a function returns ad-hoc copies of the slice in
// order to preculde the possibility of this important collection being modified
// at runtime by a client.
func WorkerPhasesNonTerminal() []WorkerPhase {
	return []WorkerPhase{
		WorkerPhasePending,
		WorkerPhaseRunning,
		WorkerPhaseUnknown,
	}
}

// WorkerSpec is the technical blueprint for a Worker.
type WorkerSpec struct {
	// Container encapsulates the details of an OCI container that forms the
	// cornerstone of the Worker.
	Container ContainerSpec `json:"container"`
	// WorkspaceSize specifies the size of a volume that will be provisioned as
	// a shared workspace for the Worker and any Jobs it spawns.
	// The value can be expressed in bytes (as a plain integer) or as a
	// fixed-point integer using one of these suffixes: E, P, T, G, M, K.
	// Power-of-two equivalents may also be used: Ei, Pi, Ti, Gi, Mi, Ki.
	WorkspaceSize string `json:"workspaceSize"`
	// Git encapsulates git-specific Worker details.
	Git WorkerGitConfig `json:"git"`
	// Kubernetes encapsulates Kubernetes-specific Worker details.
	Kubernetes WorkerKubernetesConfig `json:"kubernetes"`
	// JobsSpec encapsulates configuration and policies for any Jobs spawned by
	// the Worker.
	Jobs JobsSpec `json:"jobs"`
	// LogLevel specifies the desired granularity of Worker log output.
	LogLevel LogLevel `json:"logLevel"`
	// ConfigFilesDirectory specifies a directory within the Worker's workspace
	// where any relevant configuration files (e.g. brigade.json, brigade.js,
	// etc.) can be located.
	ConfigFilesDirectory string `json:"configFilesDirectory"`
	// DefaultConfigFiles is a map of configuration file names to configuration
	// file content. This is useful for Workers that do not integrate with any
	// source control system and would like to embed configuration (e.g.
	// brigade.json) or scripts (e.g. brigade.js) directly within the WorkerSpec.
	DefaultConfigFiles map[string]string `json:"defaultConfigFiles"`
}

// ContainerSpec represents the technical details of an OCI container.
type ContainerSpec struct {
	// Image specified the OCI image on which the container should be based.
	Image string `json:"image"`
	// ImagePullPolicy specifies whether a container host already having the
	// specified OCI image should attempt to re-pull that image prior to launching
	// a new container.
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy"`
	// Command specifies the command to be executed by the OCI container. This
	// can be used to optionally override the default command specified by the OCI
	// image itself.
	Command []string `json:"command"`
	// Arguments specifies arguments to the command executed by the OCI container.
	Arguments []string `json:"arguments"`
	// Environment is a map of key/value pairs that specify environment variables
	// to be set within the OCI container.
	Environment map[string]string `json:"environment"`
}

// WorkerGitConfig represents git-specific Worker details.
type WorkerGitConfig struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL"`
	// Commit specifies a commit (by sha) to be checked out.
	Commit string `json:"commit"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// defualt to "master" at runtime.
	Ref string `json:"ref"`
	// InitSubmodules indicates whether to clone the repository's submodules.
	InitSubmodules bool `json:"initSubmodules"`
}

// WorkerKubernetesConfig represents Kubernetes-specific Worker configuration.
type WorkerKubernetesConfig struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which the Worker's container is based. The
	// default worker image is publicly available on Docker Hub and as such this
	// field only needs to be utilized in the case of private, custom worker
	// images. The image pull secrets in question must be created out-of-band by a
	// sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// WorkerStatus represents the status of a Worker.
type WorkerStatus struct {
	// Started indicates the time the Worker began execution. It will be nil for
	// a Worker that is not yet executing.
	Started *time.Time `json:"started"`
	// Started indicates the time the Worker concluded execution. It will be nil
	// for a Worker that is not done executing (or hasn't started).
	Ended *time.Time `json:"ended"`
	// Phase indicates where the Worker is in its lifecycle.
	Phase WorkerPhase `json:"phase"`
}

// MarshalJSON amends WorkerStatus instances with type metadata so that clients
// do not need to be concerned with the tedium of doing so.
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
