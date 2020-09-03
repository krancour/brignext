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

// TODO: Document this
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string
}

type LogStreamOptions struct {
	Follow bool `json:"follow"`
}

// LogsClient is the specialized client for managing Event Logs with the
// BrigNext API.
type LogsClient interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or a specific
	// container thereof), are streamed.
	Stream(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan sdk.LogEntry, <-chan error, error)
}

type logsClient struct {
	*baseClient
}

// NewLogsClient returns a specialized client for managing Event Logs.
func NewLogsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) LogsClient {
	return &logsClient{
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

func (l *logsClient) Stream(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan sdk.LogEntry, <-chan error, error) {
	queryParams := map[string]string{}
	if selector.Job != "" {
		queryParams["job"] = selector.Job
	}
	if selector.Container != "" {
		queryParams["container"] = selector.Container
	}
	if opts.Follow {
		queryParams["follow"] = "true"
	}

	resp, err := l.submitRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			authHeaders: l.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan sdk.LogEntry)
	errCh := make(chan error)

	go l.receiveStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

// receiveStream is used to receive log messages as SSEs (server sent events),
// decode those, and publish them to a channel.
func (l *logsClient) receiveStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- sdk.LogEntry,
	errCh chan<- error,
) {
	defer close(logEntryCh)
	defer close(errCh)
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := sdk.LogEntry{}
		if err := decoder.Decode(&logEntry); err != nil {
			if err == io.EOF {
				return
			}
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case logEntryCh <- logEntry:
		case <-ctx.Done():
			return
		}
	}
}
