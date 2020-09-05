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

// WorkersClient is the specialized client for managing Event Worker with the
// Brigade API.
type WorkersClient interface {
	// Start starts the indicated Event's Worker on Brigade's workload execution
	// substrate.
	Start(ctx context.Context, eventID string) error
	// Get returns an Event's Worker's status.
	GetStatus(ctx context.Context, eventID string) (core.WorkerStatus, error)
	// WatchStatus returns a channel over which an Event's Worker's status is
	// streamed. The channel receives a new WorkerStatus every time there is any
	// change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
	) (<-chan core.WorkerStatus, <-chan error, error)
	// UpdateStatus updates the status of an Event's Worker.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status core.WorkerStatus,
	) error

	Jobs() JobsClient
}

type workersClient struct {
	*restmachinery.BaseClient
	jobsClient JobsClient
}

// NewWorkersClient returns a specialized client for managing Event Workers.
func NewWorkersClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) WorkersClient {
	return &workersClient{
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
		jobsClient: NewJobsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (w *workersClient) Start(ctx context.Context, eventID string) error {
	return w.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/worker/start", eventID),
			AuthHeaders: w.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (w *workersClient) GetStatus(
	ctx context.Context,
	eventID string,
) (core.WorkerStatus, error) {
	status := core.WorkerStatus{}
	return status, w.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			AuthHeaders: w.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &status,
		},
	)
}

func (w *workersClient) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan core.WorkerStatus, <-chan error, error) {
	resp, err := w.SubmitRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			AuthHeaders: w.BearerTokenAuthHeaders(),
			QueryParams: map[string]string{
				"watch": "true",
			},
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	statusCh := make(chan core.WorkerStatus)
	errCh := make(chan error)

	go w.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (w *workersClient) UpdateStatus(
	_ context.Context,
	eventID string,
	status core.WorkerStatus,
) error {
	return w.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			AuthHeaders: w.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (w *workersClient) Jobs() JobsClient {
	return w.jobsClient
}

func (w *workersClient) receiveStatusStream(
	ctx context.Context,
	reader io.ReadCloser,
	statusCh chan<- core.WorkerStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := core.WorkerStatus{}
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
