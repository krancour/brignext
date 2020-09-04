package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/core"
	"github.com/krancour/brignext/v2/sdk/internal/restmachinery"
)

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
	) (<-chan core.LogEntry, <-chan error, error)
}

type logsClient struct {
	*restmachinery.BaseClient
}

// NewLogsClient returns a specialized client for managing Event Logs.
func NewLogsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) LogsClient {
	return &logsClient{
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

func (l *logsClient) Stream(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan core.LogEntry, <-chan error, error) {
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

	resp, err := l.SubmitRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: l.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan core.LogEntry)
	errCh := make(chan error)

	go l.receiveStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

// receiveStream is used to receive log messages as SSEs (server sent events),
// decode those, and publish them to a channel.
func (l *logsClient) receiveStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- core.LogEntry,
	errCh chan<- error,
) {
	defer close(logEntryCh)
	defer close(errCh)
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := core.LogEntry{}
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
