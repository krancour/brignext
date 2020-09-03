package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
)

// JobsClient is the specialized client for managing Event Jobs with the
// BrigNext API.
type JobsClient interface {
	// Create, given an Event identifier and JobSpec, creates a new pending Job
	// and schedules it for execution.
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec sdk.JobSpec,
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
	) (sdk.JobStatus, error)
	// WatchStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan sdk.JobStatus, <-chan error, error)
	// UpdateStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status sdk.JobStatus,
	) error
}

type jobsClient struct {
	*baseClient
}

// NewJobsClient returns a specialized client for managing Event Jobs.
func NewJobsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) JobsClient {
	return &jobsClient{
		baseClient: &baseClient{
			apiAddress: apiAddress,
			apiToken:   apiToken,
			httpClient: &http.Client{
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
	jobSpec sdk.JobSpec,
) error {
	return j.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/spec",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			reqBodyObj:  jobSpec,
			successCode: http.StatusCreated,
		},
	)
}

func (j *jobsClient) Start(
	ctx context.Context,
	eventID string,
	jobName string,
) error {
	return j.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/start",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) GetStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (sdk.JobStatus, error) {
	status := sdk.JobStatus{}
	return status, j.executeRequest(
		outboundRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &status,
		},
	)
}

func (j *jobsClient) WatchStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan sdk.JobStatus, <-chan error, error) {
	resp, err := j.submitRequest(
		outboundRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"watch": "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	statusCh := make(chan sdk.JobStatus)
	errCh := make(chan error)

	go j.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (j *jobsClient) UpdateStatus(
	_ context.Context,
	eventID string,
	jobName string,
	status sdk.JobStatus,
) error {
	return j.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
		},
	)
}

func (j *jobsClient) receiveStatusStream(
	ctx context.Context,
	reader io.ReadCloser,
	statusCh chan<- sdk.JobStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := sdk.JobStatus{}
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
