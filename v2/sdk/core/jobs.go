package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
	"github.com/brigadecore/brigade/v2/sdk/meta"
)

// JobPhase represents where a Job is within its lifecycle.
type JobPhase string

const (
	// JobPhaseAborted represents the state wherein a Job was forcefully
	// stopped during execution.
	JobPhaseAborted JobPhase = "ABORTED"
	// JobPhaseFailed represents the state wherein a Job has run to
	// completion but experienced errors.
	JobPhaseFailed JobPhase = "FAILED"
	// JobPhasePending represents the state wherein a Job is awaiting
	// execution.
	JobPhasePending JobPhase = "PENDING"
	// JobPhaseRunning represents the state wherein a Job is currently
	// being executed.
	JobPhaseRunning JobPhase = "RUNNING"
	// JobPhaseSucceeded represents the state where a Job has run to
	// completion without error.
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
	// JobPhaseTimedOut represents the state wherein a Job has has not completed
	// within a designated timeframe.
	JobPhaseTimedOut JobPhase = "TIMED_OUT"
	// JobPhaseUnknown represents the state wherein a Job's state is unknown. Note
	// that this is possible if and only if the underlying Job execution substrate
	// (Kubernetes), for some unanticipated, reason does not know the Job's
	// (Pod's) state.
	JobPhaseUnknown WorkerPhase = "UNKNOWN"
)

// Job represents a component spawned by a Worker to complete a single task
// during the handling of an Event.
type Job struct {
	// Spec is the technical blueprint for the Job.
	Spec JobSpec `json:"spec"`
	// Status contains details of the Job's current state.
	Status JobStatus `json:"status"`
}

// JobSpec is the technical blueprint for a Job.
type JobSpec struct {
	// PrimaryContainer specifies the details of an OCI container that forms the
	// cornerstone of the Job. Job success or failure is tied to completion and
	// exit code of this container.
	PrimaryContainer JobContainerSpec `json:"primaryContainer"`
	// SidecarContainers specifies the details of supplemental, "sidecar"
	// containers. Their completion and exit code do not directly impact Job
	// status. Brigade does not understand dependencies between a Job's multiple
	// containers and cannot enforce any specific startup or shutdown order. When
	// such dependencies exist (for instance, a primary container than cannot
	// proceed with a suite of tests until a database is launched and READY in a
	// sidecar container), then logic within those containers must account for
	// these constraints.
	SidecarContainers map[string]JobContainerSpec `json:"sidecarContainers,omitempty"` // nolint: lll
	TimeoutSeconds    int64                       `json:"timeoutSeconds,omitempty"`
	Host              *JobHost                    `json:"host,omitempty"`
}

// MarshalJSON amends JobSpec instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (j JobSpec) MarshalJSON() ([]byte, error) {
	type Alias JobSpec
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "JobSpec",
			},
			Alias: (Alias)(j),
		},
	)
}

type JobHost struct {
	OS           string            `json:"os,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// JobStatus represents the status of a Job.
type JobStatus struct {
	// Started indicates the time the Job began execution.
	Started *time.Time `json:"started,omitempty"`
	// Ended indicates the time the Job concluded execution. It will be nil
	// for a Job that is not done executing.
	Ended *time.Time `json:"ended,omitempty"`
	// Phase indicates where the Job is in its lifecycle.
	Phase JobPhase `json:"phase,omitempty"`
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

// JobsClient is the specialized client for managing Event Jobs with the
// Brigade API.
type JobsClient interface {
	// Create, given an Event identifier and JobSpec, creates a new pending Job
	// and schedules it for execution.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	// Start initiates execution of a pending Job.
	Start(
		ctx context.Context,
		eventID string,
		jobName string,
	) error
	// GetStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan JobStatus, <-chan error, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}

type jobsClient struct {
	*restmachinery.BaseClient
}

// NewJobsClient returns a specialized client for managing Event Jobs.
func NewJobsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) JobsClient {
	return &jobsClient{
		BaseClient: &restmachinery.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (j *jobsClient) Create(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec JobSpec,
) error {
	return j.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/spec",
				eventID,
				jobName,
			),
			AuthHeaders: j.BearerTokenAuthHeaders(),
			ReqBodyObj:  jobSpec,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (j *jobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	return j.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/start",
				eventID,
				jobName,
			),
			AuthHeaders: j.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (JobStatus, error) {
	status := JobStatus{}
	return status, j.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method: http.MethodGet,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			AuthHeaders: j.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &status,
		},
	)
}

func (j *jobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan JobStatus, <-chan error, error) {
	resp, err := j.SubmitRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method: http.MethodGet,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			AuthHeaders: j.BearerTokenAuthHeaders(),
			QueryParams: map[string]string{
				"watch": "true",
			},
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	statusCh := make(chan JobStatus)
	errCh := make(chan error)

	go j.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (j *jobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	return j.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			AuthHeaders: j.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) receiveStatusStream(
	ctx context.Context,
	reader io.ReadCloser,
	statusCh chan<- JobStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := JobStatus{}
		if err := decoder.Decode(&status); err != nil {
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case statusCh <- status:
		case <-ctx.Done():
			return
		}
	}
}