package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/brigadecore/brigade/v2/sdk/core"
	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
)

// JobsClient is the specialized client for managing Event Jobs with the
// Brigade API.
type JobsClient interface {
	// Create, given an Event identifier and JobSpec, creates a new pending Job
	// and schedules it for execution.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec core.JobSpec,
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
	) (core.JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan core.JobStatus, <-chan error, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status core.JobStatus,
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
	jobSpec core.JobSpec,
) error {
	return j.ExecuteRequest(
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
) (core.JobStatus, error) {
	status := core.JobStatus{}
	return status, j.ExecuteRequest(
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
) (<-chan core.JobStatus, <-chan error, error) {
	resp, err := j.SubmitRequest(
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

	statusCh := make(chan core.JobStatus)
	errCh := make(chan error)

	go j.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (j *jobsClient) UpdateStatus(
	_ context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
) error {
	return j.ExecuteRequest(
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
	statusCh chan<- core.JobStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := core.JobStatus{}
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
