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
}

// EventReference is an abridged representation of an Event useful to
// API operations that construct and return potentially large collections of
// events. Utilizing such an abridged representation limits response size
// significantly as Events have the potentia to be quite large.
type EventReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	ProjectID                string          `json:"projectID,omitempty"`
	Source                   string          `json:"source,omitempty"`
	Type                     string          `json:"type,omitempty"`
	WorkerPhase              sdk.WorkerPhase `json:"workerPhase,omitempty"`
}

// MarshalJSON amends EventReference instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (e EventReference) MarshalJSON() ([]byte, error) {
	type Alias EventReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReference",
			},
			Alias: (Alias)(e),
		},
	)
}

// EventReferenceList is an ordered list of EventtReferences.
type EventReferenceList struct {
	// TODO: When pagination is implemented, list metadata will need to be added
	// Items is a slice of EventReferences.
	Items []EventReference `json:"items,omitempty"`
}

// MarshalJSON amends EventReferenceList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (e EventReferenceList) MarshalJSON() ([]byte, error) {
	type Alias EventReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReferenceList",
			},
			Alias: (Alias)(e),
		},
	)
}

// LogOptions represents useful criteria for identifying a specific container
// of a specific Job when requesting Event logs.
type LogOptions struct {
	Job       string `json:"job,omitempty"`
	Container string `json:"container,omitempty"`
}

// EventsClient is the specialized interface for managing Events with the
// BrigNext API.
type EventsClient interface {
	Create(context.Context, sdk.Event) (EventReferenceList, error)
	List(
		context.Context,
		EventListOptions,
	) (EventReferenceList, error)
	Get(context.Context, string) (sdk.Event, error)
	Cancel(context.Context, string) error
	CancelCollection(
		context.Context,
		EventListOptions,
	) (EventReferenceList, error)
	Delete(context.Context, string) error
	DeleteCollection(
		context.Context,
		EventListOptions,
	) (EventReferenceList, error)

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
		opts LogOptions,
	) (sdk.LogEntryList, error)
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
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  event,
			successCode: http.StatusCreated,
			respObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) List(
	_ context.Context,
	opts EventListOptions,
) (EventReferenceList, error) {
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
	eventList := EventReferenceList{}
	return eventList, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
			respObj:     &eventList,
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

func (e *eventsClient) CancelCollection(
	_ context.Context,
	opts EventListOptions,
) (EventReferenceList, error) {
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
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/events/cancellations",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
			respObj:     &eventRefList,
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

func (e *eventsClient) DeleteCollection(
	_ context.Context,
	opts EventListOptions,
) (EventReferenceList, error) {
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
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
			respObj:     &eventRefList,
		},
	)
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
	logEntryList := sdk.LogEntryList{}
	return logEntryList, e.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: e.queryParamsFromLogOptions(opts, false), // Don't stream
			successCode: http.StatusOK,
			respObj:     &logEntryList,
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
