package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// LogLevel represents the desired granularity of Worker log output.
type LogLevel string

// LogLevelInfo represents INFO level granularity in Worker log output.
const LogLevelInfo LogLevel = "INFO"

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
// returns ad-hoc copies of the slice in order to preclude the possibility of
// this important collection being modified at runtime.
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

// Worker represents a component that orchestrates handling of a single Event.
type Worker struct {
	// Spec is the technical blueprint for the Worker.
	Spec WorkerSpec `json:"spec" bson:"spec"`
	// Status contains details of the Worker's current state.
	Status WorkerStatus `json:"status" bson:"status"`
	// Token is an API token that grants a Worker permission to create new Jobs
	// only for the Event to which it belongs.
	Token string `json:"-" bson:"-"`
	// HashedToken is a secure hash of the Token field.
	HashedToken string `json:"-" bson:"hashedToken"`
	// Jobs contains details of all Jobs spawned by the Worker during handling of
	// the Event.
	Jobs map[string]Job `json:"jobs,omitempty" bson:"jobs,omitempty"`
}

// WorkerSpec is the technical blueprint for a Worker.
// nolint: lll
type WorkerSpec struct {
	// Container specifies the details of an OCI container that forms the
	// cornerstone of the Worker.
	Container *ContainerSpec `json:"container,omitempty" bson:"container,omitempty"`
	// UseWorkspace indicates whether the Worker requires a volume to be
	// provisioned to be shared by itself and any Jobs it creates. This is a
	// generally useful feature, but by opting out of it (or rather, not
	// opting-in), Job results can be made cacheable and Jobs
	// resumable/retriable-- something which cannot be done otherwise since
	// managing the state of the shared volume would require a layered file system
	// that we currently do not have.
	UseWorkspace bool `json:"useWorkspace" bson:"useWorkspace"`
	// WorkspaceSize specifies the size of a volume that will be provisioned as
	// a shared workspace for the Worker and any Jobs it spawns.
	// The value can be expressed in bytes (as a plain integer) or as a
	// fixed-point integer using one of these suffixes: E, P, T, G, M, K.
	// Power-of-two equivalents may also be used: Ei, Pi, Ti, Gi, Mi, Ki.
	WorkspaceSize string `json:"workspaceSize,omitempty" bson:"workspaceSize,omitempty"`
	// Git contains git-specific Worker details.
	Git *WorkerGitConfig `json:"git,omitempty" bson:"git,omitempty"`
	// Kubernetes contains Kubernetes-specific Worker details.
	Kubernetes *WorkerKubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"`
	// JobPolicies specifies policies for any Jobs spawned by the Worker.
	JobPolicies *JobPolicies `json:"jobPolicies,omitempty" bson:"jobPolicies,omitempty"`
	// LogLevel specifies the desired granularity of Worker log output.
	LogLevel LogLevel `json:"logLevel,omitempty" bson:"logLevel,omitempty"`
	// ConfigFilesDirectory specifies a directory within the Worker's workspace
	// where any relevant configuration files (e.g. brigade.json, brigade.js,
	// etc.) can be located.
	ConfigFilesDirectory string `json:"configFilesDirectory,omitempty" bson:"configFilesDirectory,omitempty"`
	// DefaultConfigFiles is a map of configuration file names to configuration
	// file content. This is useful for Workers that do not integrate with any
	// source control system and would like to embed configuration (e.g.
	// brigade.json) or scripts (e.g. brigade.js) directly within the WorkerSpec.
	DefaultConfigFiles map[string]string `json:"defaultConfigFiles,omitempty" bson:"defaultConfigFiles,omitempty"`
}

// WorkerGitConfig represents git-specific Worker details.
type WorkerGitConfig struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	// Commit specifies a commit (by SHA) to be checked out.
	Commit string `json:"commit,omitempty" bson:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
	Ref string `json:"ref,omitempty" bson:"ref,omitempty"`
	// InitSubmodules indicates whether to clone the repository's submodules.
	InitSubmodules bool `json:"initSubmodules" bson:"initSubmodules"`
}

// WorkerKubernetesConfig represents Kubernetes-specific Worker configuration.
type WorkerKubernetesConfig struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which the Worker's container is based. The
	// default worker image is publicly available on Docker Hub and as such this
	// field only needs to be utilized in the case of private, custom worker
	// images. The image pull secrets in question must be created out-of-band by a
	// sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"` // nolint: lll
}

// JobPolicies represents policies for any Jobs spawned by a Worker.
type JobPolicies struct {
	// AllowPrivileged specifies whether the Worker is permitted to launch Jobs
	// that utilize privileged containers.
	AllowPrivileged bool `json:"allowPrivileged" bson:"allowPrivileged"`
	// AllowDockerSocketMount specifies whether the Worker is permitted to launch
	// Jobs that mount the underlying host's Docker socket into its own file
	// system.
	AllowDockerSocketMount bool `json:"allowDockerSocketMount" bson:"allowDockerSocketMount"` // nolint: lll
	// Kubernetes specifies Kubernetes-specific policies for any Jobs spawned by
	// the Worker.
	Kubernetes *KubernetesJobPolicies `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
}

// KubernetesJobPolicies represents Kubernetes-specific policies for any Jobs
// spawned by a Worker.
type KubernetesJobPolicies struct {
	// ImagePullSecrets enumerates any image pull secrets that Kubernetes may use
	// when pulling the OCI image on which the Jobs' containers are based. The
	// image pull secrets in question must be created out-of-band by a
	// sufficiently authorized user of the Kubernetes cluster.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty" bson:"imagePullSecrets,omitempty"` // nolint: lll
}

// WorkerStatus represents the status of a Worker.
type WorkerStatus struct {
	// Started indicates the time the Worker began execution. It will be nil for
	// a Worker that is not yet executing.
	Started *time.Time `json:"started,omitempty" bson:"started,omitempty"`
	// Ended indicates the time the Worker concluded execution. It will be nil
	// for a Worker that is not done executing (or hasn't started).
	Ended *time.Time `json:"ended,omitempty" bson:"ended,omitempty"`
	// Phase indicates where the Worker is in its lifecycle.
	Phase WorkerPhase `json:"phase,omitempty" bson:"phase,omitempty"`
}

// TODO: We probably don't need this interface. The idea is to have a single
// implementation of the service's logic, with only underlying components being
// pluggable. BUT, STRONGLY CONSIDER THAT WE MAY NEED THIS TO MOCK OUT THE
// SERVICE WHEN TESTING THE CORRESPONDING ENDPOINTS.
type WorkersService interface {
	// Start starts the indicated Event's Worker on Brigade's workload
	// execution substrate.
	Start(ctx context.Context, eventID string) error
	// GetStatus returns an Event's Worker's status.
	GetStatus(
		ctx context.Context,
		eventID string,
	) (WorkerStatus, error)
	// WatchStatus returns a channel over which an Event's Worker's status
	// is streamed. The channel receives a new WorkerStatus every time there is
	// any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
	) (<-chan WorkerStatus, error)
	// UpdateStatus updates the status of an Event's Worker.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}

type workersService struct {
	authorize    authx.AuthorizeFn
	eventsStore  EventsStore
	workersStore WorkersStore
	substrate    Substrate
}

// TODO: There probably isn't any good reason to actually have this
// constructor-like function here. Let's consider removing it.
func NewWorkersService(
	eventsStore EventsStore,
	workersStore WorkersStore,
	substrate Substrate,
) WorkersService {
	return &workersService{
		authorize:    authx.Authorize,
		eventsStore:  eventsStore,
		workersStore: workersStore,
		substrate:    substrate,
	}
}

func (w *workersService) Start(ctx context.Context, eventID string) error {
	if err := w.authorize(ctx, authx.RoleScheduler()); err != nil {
		return err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	if event.Worker.Status.Phase != WorkerPhasePending {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   event.ID,
			Reason: fmt.Sprintf(
				"Event %q worker has already been started.",
				event.ID,
			),
		}
	}

	if err = w.substrate.StartWorker(ctx, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}

func (w *workersService) GetStatus(
	ctx context.Context,
	eventID string,
) (WorkerStatus, error) {
	if err := w.authorize(ctx, authx.RoleReader()); err != nil {
		return WorkerStatus{}, err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return WorkerStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	return event.Worker.Status, nil
}

// TODO: Should we put some kind of timeout on this function so forgetful
// clients cannot watch forever?
func (w *workersService) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan WorkerStatus, error) {
	if err := w.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event up front to confirm it exists.
	if _, err := w.eventsStore.Get(ctx, eventID); err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	statusCh := make(chan WorkerStatus)
	go func() {
		defer close(statusCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			event, err := w.eventsStore.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- event.Worker.Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (w *workersService) UpdateStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	if err := w.authorize(ctx, authx.RoleObserver()); err != nil {
		return err
	}

	if err := w.workersStore.UpdateStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}
	return nil
}

type WorkersStore interface {
	UpdateSpec(
		ctx context.Context,
		eventID string,
		spec WorkerSpec,
	) error
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}
