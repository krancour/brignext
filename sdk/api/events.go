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
	"github.com/krancour/brignext/v2/sdk/meta"
)

// EventListOptions represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventListOptions struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Worker's in any of the
	// indicated phases should be selected.
	WorkerPhases []sdk.WorkerPhase
	// Continue aids in pagination of long lists. It permits clients to echo an
	// opaque value obtained from a previous API call back to the API in a
	// subsequent call in order to indicate what resource was the last on the
	// previous page.
	Continue string
	// Limit aids in pagination of long lists. It permits clients to specify page
	// size when making API calls. The API server provides a default when a value
	// is not specified and may reject or override invalid values (non-positive)
	// numbers or very large page sizes.
	Limit int64
}

// EventList is an ordered and pageable list of Events.
type EventList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Events.
	Items []sdk.Event `json:"items,omitempty"`
}

// MarshalJSON amends EventList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (e EventList) MarshalJSON() ([]byte, error) {
	type Alias EventList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventList",
			},
			Alias: (Alias)(e),
		},
	)
}

type CancelManyEventsResult struct {
	Count int64 `json:count`
}

type DeleteManyEventsResult struct {
	Count int64 `json:count`
}

// LogOptions represents useful criteria for identifying a specific container
// of a specific Job when requesting Event logs.
type LogOptions struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string `json:"job,omitempty"`
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string `json:"container,omitempty"`
	// Continue aids in pagination of long lists. It permits clients to echo an
	// opaque value obtained from a previous API call back to the API in a
	// subsequent call in order to indicate what resource was the last on the
	// previous page.
	Continue string
	// Limit aids in pagination of long lists. It permits clients to specify page
	// size when making API calls. The API server provides a default when a value
	// is not specified and may reject or override invalid values (non-positive)
	// numbers or very large page sizes.
	Limit int64
}

// EventsClient is the specialized client for managing Events with the BrigNext
// API.
type EventsClient interface {
	// Create creates a new Event.
	Create(context.Context, sdk.Event) (EventList, error)
	// List returns an EventList, with its Items (Events) ordered by age, newest
	// first. Criteria for which Events should be retrieved can be specified using
	// the EventListOptions parameter.
	List(context.Context, EventListOptions) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (sdk.Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(context.Context, EventListOptions) (CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(context.Context, EventListOptions) (DeleteManyEventsResult, error)

	// StartWorker starts the indicated Event's Worker on BrigNext's worlkoad
	// execution substrate.
	StartWorker(ctx context.Context, eventID string) error
	// GetWorkerStatus returns an Event's Worker's status.
	GetWorkerStatus(ctx context.Context, eventID string) (sdk.WorkerStatus, error)
	// WatchWorkerStatus returns a channel over which an Event's Worker's status
	// is streamed. The channel receives a new WorkerStatus every time there is
	// any change in that status.
	WatchWorkerStatus(
		ctx context.Context,
		eventID string,
	) (<-chan sdk.WorkerStatus, <-chan error, error)
	// UpdateWorkerStatus updates the status of an Event's Worker.
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status sdk.WorkerStatus,
	) error

	// CreateJob, given an Event identifier and JobSpec, creates a new Job and
	// starts it on BrigNext's worlkoad execution substrate.
	CreateJob(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec sdk.JobSpec,
	) error
	// GetJobStatus, given an Event identifier and Job name, returns the Job's
	// status.
	GetJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (sdk.JobStatus, error)
	// WatchJobStatus, given an Event identifier and Job name, returns a channel
	// over which the Job's status is streamed. The channel receives a new
	// JobStatus every time there is any change in that status.
	WatchJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan sdk.JobStatus, <-chan error, error)
	// UpdateJobStatus, given an Event identifier and Job name, updates the status
	// of that Job.
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status sdk.JobStatus,
	) error

	// GetLogs retrieves logs for an Event's Worker, or using the LogOptions
	// parameter, a Job spawned by that Worker (or specific container thereof).
	GetLogs(
		ctx context.Context,
		eventID string,
		opts LogOptions,
	) (sdk.LogEntryList, error)
	// StreamLogs returns a channel over which logs for an Event's Worker, or
	// using the LogOptions parameter, a Job spawned by that Worker (or specific
	// container thereof), are streamed.
	StreamLogs(
		ctx context.Context,
		eventID string,
		opts LogOptions,
	) (<-chan sdk.LogEntry, <-chan error, error)
}

type eventsClient struct {
	*baseClient
}

// NewEventsClient returns a specialized client for managing Events.
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
) (EventList, error) {
	events := EventList{}
	return events, e.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  event,
			successCode: http.StatusCreated,
			respObj:     &events,
		},
	)
}

func (e *eventsClient) List(
	_ context.Context,
	opts EventListOptions,
) (EventList, error) {
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
	events := EventList{}
	return events, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: e.appendListQueryParams(queryParams, opts.Continue, opts.Limit),
			successCode: http.StatusOK,
			respObj:     &events,
		},
	)
}

func (e *eventsClient) Get(
	_ context.Context,
	id string,
) (sdk.Event, error) {
	event := sdk.Event{}
	return event, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &event,
		},
	)
}

func (e *eventsClient) Cancel(_ context.Context, id string) error {
	return e.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/cancellation", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CancelMany(
	_ context.Context,
	opts EventListOptions,
) (CancelManyEventsResult, error) {
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
	result := CancelManyEventsResult{}
	return result, e.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/events/cancellations",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
			respObj:     &result,
		},
	)
}

func (e *eventsClient) Delete(_ context.Context, id string) error {
	return e.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/events/%s", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) DeleteMany(
	_ context.Context,
	opts EventListOptions,
) (DeleteManyEventsResult, error) {
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
	result := DeleteManyEventsResult{}
	return result, e.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
			respObj:     &result,
		},
	)
}

func (e *eventsClient) StartWorker(ctx context.Context, eventID string) error {
	return e.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/start", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) GetWorkerStatus(
	ctx context.Context,
	eventID string,
) (sdk.WorkerStatus, error) {
	status := sdk.WorkerStatus{}
	return status, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &status,
		},
	)
}

func (e *eventsClient) WatchWorkerStatus(
	ctx context.Context,
	eventID string,
) (<-chan sdk.WorkerStatus, <-chan error, error) {
	resp, err := e.submitRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
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

	go e.receiveWorkerStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (e *eventsClient) UpdateWorkerStatus(
	_ context.Context,
	eventID string,
	status sdk.WorkerStatus,
) error {
	return e.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) CreateJob(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec sdk.JobSpec,
) error {
	return e.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/spec",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  jobSpec,
			successCode: http.StatusCreated,
		},
	)
}

func (e *eventsClient) GetJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (sdk.JobStatus, error) {
	status := sdk.JobStatus{}
	return status, e.executeRequest(
		outboundRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &status,
		},
	)
}

func (e *eventsClient) WatchJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan sdk.JobStatus, <-chan error, error) {
	resp, err := e.submitRequest(
		outboundRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
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

	go e.receiveJobStatusStream(ctx, resp.Body, statusCh, errCh)

	return statusCh, errCh, nil
}

func (e *eventsClient) UpdateJobStatus(
	_ context.Context,
	eventID string,
	jobName string,
	status sdk.JobStatus,
) error {
	return e.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) GetLogs(
	ctx context.Context,
	eventID string,
	opts LogOptions,
) (sdk.LogEntryList, error) {
	logEntries := sdk.LogEntryList{}
	return logEntries, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: e.queryParamsFromLogOptions(opts, false), // Don't stream
			successCode: http.StatusOK,
			respObj:     &logEntries,
		},
	)
}

func (e *eventsClient) StreamLogs(
	ctx context.Context,
	eventID string,
	opts LogOptions,
) (<-chan sdk.LogEntry, <-chan error, error) {
	resp, err := e.submitRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: e.queryParamsFromLogOptions(opts, true), // Stream
			successCode: http.StatusOK,
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

// queryParamsFromLogOptions creates a map[string]string of query parameters
// based on the values of each field in the provided LogOptions and a boolean
// indicating whether the client is requesting a log stream (and not a static
// list of log messages).
func (e *eventsClient) queryParamsFromLogOptions(
	opts LogOptions,
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
	return e.appendListQueryParams(queryParams, opts.Continue, opts.Limit)
}

func (e *eventsClient) receiveWorkerStatusStream(
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

func (e *eventsClient) receiveJobStatusStream(
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

// receiveLogStream is used to receive log messages as SSEs (server sent
// events), decode those, and publish them to a channel.
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
