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

// WorkersClient is the specialized client for managing Event Worker with the
// BrigNext API.
type WorkersClient interface {
	// Start starts the indicated Event's Worker on BrigNext's workload execution
	// substrate.
	Start(ctx context.Context, eventID string) error
	// Get returns an Event's Worker's status.
	GetStatus(ctx context.Context, eventID string) (sdk.WorkerStatus, error)
	// WatchStatus returns a channel over which an Event's Worker's status is
	// streamed. The channel receives a new WorkerStatus every time there is any
	// change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
	) (<-chan sdk.WorkerStatus, <-chan error, error)
	// UpdateStatus updates the status of an Event's Worker.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status sdk.WorkerStatus,
	) error

	Jobs() JobsClient
}

type workersClient struct {
	*baseClient
	jobsClient JobsClient
}

// NewWorkersClient returns a specialized client for managing Event Workers.
func NewWorkersClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) WorkersClient {
	return &workersClient{
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
		jobsClient: NewJobsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (w *workersClient) Start(ctx context.Context, eventID string) error {
	return w.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/start", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (w *workersClient) GetStatus(
	ctx context.Context,
	eventID string,
) (sdk.WorkerStatus, error) {
	status := sdk.WorkerStatus{}
	return status, w.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &status,
		},
	)
}

func (w *workersClient) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan sdk.WorkerStatus, <-chan error, error) {
	resp, err := w.submitRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"watch": "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	statusCh := make(chan sdk.WorkerStatus)
	errCh := make(chan error)

	go w.receiveStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (w *workersClient) UpdateStatus(
	_ context.Context,
	eventID string,
	status sdk.WorkerStatus,
) error {
	return w.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
		},
	)
}

func (w *workersClient) Jobs() JobsClient {
	return w.jobsClient
}

func (w *workersClient) receiveStatusStream(
	ctx context.Context,
	reader io.ReadCloser,
	statusCh chan<- sdk.WorkerStatus,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		status := sdk.WorkerStatus{}
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
