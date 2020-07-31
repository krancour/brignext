package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
)

type EventsClient interface {
	Create(context.Context, brignext.Event) (brignext.EventReferenceList, error)
	List(
		context.Context,
		brignext.EventListOptions,
	) (brignext.EventReferenceList, error)
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

type eventsClient struct {
	*apimachinery.BaseClient
}

func NewEventsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) EventsClient {
	return &eventsClient{
		BaseClient: &apimachinery.BaseClient{
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

func (e *eventsClient) Create(
	_ context.Context,
	event brignext.Event,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{}
	return eventRefList, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			ReqBodyObj:  event,
			SuccessCode: http.StatusCreated,
			RespObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) List(
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
	eventList := brignext.EventReferenceList{}
	return eventList, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventList,
		},
	)
}

func (e *eventsClient) Get(
	_ context.Context,
	id string,
) (brignext.Event, error) {
	event := brignext.Event{}
	return event, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &event,
		},
	)
}

func (e *eventsClient) Cancel(_ context.Context, id string) error {
	return e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CancelCollection(
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
	return eventRefList, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events/cancellations",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) Delete(_ context.Context, id string) error {
	return e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) DeleteCollection(
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
	return eventRefList, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) UpdateWorkerStatus(
	_ context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	return e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) UpdateJobStatus(
	_ context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	return e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			ReqBodyObj:  status,
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) GetLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (brignext.LogEntryList, error) {
	logEntryList := brignext.LogEntryList{}
	return logEntryList, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: e.queryParamsFromLogOptions(opts, false), // Don't stream
			SuccessCode: http.StatusOK,
			RespObj:     &logEntryList,
		},
	)
}

func (e *eventsClient) StreamLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (<-chan brignext.LogEntry, <-chan error, error) {
	resp, err := e.SubmitRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: e.queryParamsFromLogOptions(opts, true), // Stream
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan brignext.LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (e *eventsClient) queryParamsFromLogOptions(
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

func (e *eventsClient) receiveLogStream(
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
