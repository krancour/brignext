package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type WorkersClient interface {
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
	GetLogs(ctx context.Context, eventID string) (LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
	) (<-chan LogEntry, <-chan error, error)
	GetInitLogs(
		ctx context.Context,
		eventID string,
	) (LogEntryList, error)
	StreamInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan LogEntry, <-chan error, error)
}

type workersClient struct {
	*baseClient
}

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
	}
}

func (w *workersClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	return w.doAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
}

func (w *workersClient) GetLogs(
	ctx context.Context,
	eventID string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	err := w.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &logEntryList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
	return logEntryList, err
}

func (w *workersClient) StreamLogs(
	ctx context.Context,
	eventID string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := w.doAPIRequest2(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
			},
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go w.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (w *workersClient) GetInitLogs(
	ctx context.Context,
	eventID string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	err := w.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"init": "true",
			},
			successCode: http.StatusOK,
			respObj:     &logEntryList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
	return logEntryList, err
}

func (w *workersClient) StreamInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := w.doAPIRequest2(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: w.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
				"init":   "true",
			},
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go w.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}
