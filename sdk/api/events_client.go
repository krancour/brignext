package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krancour/brignext/v2/sdk"
)

type EventsClient interface {
	Create(context.Context, sdk.Event) (sdk.EventReferenceList, error)
	List(
		context.Context,
		sdk.EventListOptions,
	) (sdk.EventReferenceList, error)
	Get(context.Context, string) (sdk.Event, error)
	Cancel(context.Context, string) error
	CancelCollection(
		context.Context,
		sdk.EventListOptions,
	) (sdk.EventReferenceList, error)
	Delete(context.Context, string) error
	DeleteCollection(
		context.Context,
		sdk.EventListOptions,
	) (sdk.EventReferenceList, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status sdk.WorkerStatus,
	) error

	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status sdk.JobStatus,
	) error

	GetLogs(
		ctx context.Context,
		eventID string,
		opts sdk.LogOptions,
	) (sdk.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
		opts sdk.LogOptions,
	) (<-chan sdk.LogEntry, <-chan error, error)
}

type eventsClient struct {
	*baseClient
}

func NewEventsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) EventsClient {
	return &eventsClient{
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

func (e *eventsClient) Create(
	_ context.Context,
	event sdk.Event,
) (sdk.EventReferenceList, error) {
	eventRefList := sdk.EventReferenceList{}
	return eventRefList, e.ExecuteRequest(
		OutboundRequest{
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
	opts sdk.EventListOptions,
) (sdk.EventReferenceList, error) {
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
	eventList := sdk.EventReferenceList{}
	return eventList, e.ExecuteRequest(
		OutboundRequest{
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
) (sdk.Event, error) {
	event := sdk.Event{}
	return event, e.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CancelCollection(
	_ context.Context,
	opts sdk.EventListOptions,
) (sdk.EventReferenceList, error) {
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
	eventRefList := sdk.EventReferenceList{}
	return eventRefList, e.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/events/%s", id),
			AuthHeaders: e.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) DeleteCollection(
	_ context.Context,
	opts sdk.EventListOptions,
) (sdk.EventReferenceList, error) {
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
	eventRefList := sdk.EventReferenceList{}
	return eventRefList, e.ExecuteRequest(
		OutboundRequest{
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
	status sdk.WorkerStatus,
) error {
	return e.ExecuteRequest(
		OutboundRequest{
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
	status sdk.JobStatus,
) error {
	return e.ExecuteRequest(
		OutboundRequest{
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
	opts sdk.LogOptions,
) (sdk.LogEntryList, error) {
	logEntryList := sdk.LogEntryList{}
	return logEntryList, e.ExecuteRequest(
		OutboundRequest{
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
	opts sdk.LogOptions,
) (<-chan sdk.LogEntry, <-chan error, error) {
	resp, err := e.SubmitRequest(
		OutboundRequest{
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

	logCh := make(chan sdk.LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (e *eventsClient) queryParamsFromLogOptions(
	opts sdk.LogOptions,
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
	logEntryCh chan<- sdk.LogEntry,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := sdk.LogEntry{}
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
