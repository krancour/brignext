package events

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type Client interface {
	Create(context.Context, brignext.Event) (brignext.EventReferenceList, error)
	List(context.Context, brignext.EventListOptions) (brignext.EventList, error)
	Get(context.Context, string) (brignext.Event, error)
	Cancel(context.Context, string) error
	CancelCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
	Delete(context.Context, string) error
	DeleteCollection(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error

	GetLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (brignext.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (<-chan brignext.LogEntry, <-chan error, error)
}

type client struct {
	*api.BaseClient
}

func NewClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) Client {
	return &client{
		BaseClient: &api.BaseClient{
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

func (c *client) Create(
	_ context.Context,
	event brignext.Event,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{}
	return eventRefList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  event,
			SuccessCode: http.StatusCreated,
			RespObj:     &eventRefList,
		},
	)
}

func (c *client) List(
	_ context.Context,
	opts brignext.EventListOptions,
) (brignext.EventList, error) {
	queryParams := map[string]string{}
	if opts.ProjectID != "" {
		queryParams["projectID"] = opts.ProjectID
	}
	if len(opts.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(opts.WorkerPhases))
		for i, workerPhase := range opts.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	eventList := brignext.EventList{}
	return eventList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/events",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventList,
		},
	)
}

func (c *client) Get(_ context.Context, id string) (brignext.Event, error) {
	event := brignext.Event{}
	return event, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &event,
		},
	)
}

func (c *client) Cancel(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) CancelCollection(
	_ context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	queryParams := map[string]string{}
	if opts.ProjectID != "" {
		queryParams["projectID"] = opts.ProjectID
	}
	if len(opts.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(opts.WorkerPhases))
		for i, workerPhase := range opts.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	eventRefList := brignext.EventReferenceList{}
	return eventRefList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events/cancellations",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventRefList,
		},
	)
}

func (c *client) Delete(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) DeleteCollection(
	_ context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	queryParams := map[string]string{}
	if opts.ProjectID != "" {
		queryParams["projectID"] = opts.ProjectID
	}
	if len(opts.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(opts.WorkerPhases))
		for i, workerPhase := range opts.WorkerPhases {
			workerPhaseStrs[i] = string(workerPhase)
		}
		queryParams["workerPhases"] = strings.Join(workerPhaseStrs, ",")
	}
	eventRefList := brignext.EventReferenceList{}
	return eventRefList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/events",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventRefList,
		},
	)
}

func (c *client) UpdateWorkerStatus(
	_ context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) UpdateJobStatus(
	_ context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) GetLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	return logEntryList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			QueryParams: c.queryParamsFromLogOptions(opts, false), // Don't stream
			SuccessCode: http.StatusOK,
			RespObj:     &logEntryList,
		},
	)
}

func (c *client) StreamLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (<-chan brignext.LogEntry, <-chan error, error) {
	resp, err := c.SubmitRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			QueryParams: c.queryParamsFromLogOptions(opts, true), // Stream
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan brignext.LogEntry)
	errCh := make(chan error)

	go c.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (c *client) queryParamsFromLogOptions(
	opts brignext.LogOptions,
	stream bool,
) map[string]string {
	queryParams := map[string]string{}
	if opts.Job != "" {
		queryParams["job"] = opts.Job
	}
	if opts.Container != "" {
		queryParams["container"] = opts.Container
	}
	if stream {
		queryParams["stream"] = "true"
	}
	return queryParams
}

func (c *client) receiveLogStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- brignext.LogEntry,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := brignext.LogEntry{}
		if err := decoder.Decode(&logEntry); err != nil {
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
