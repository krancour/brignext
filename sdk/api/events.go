package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// EventsSelector represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventsSelector struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Worker's in any of the
	// indicated phases should be selected.
	WorkerPhases []sdk.WorkerPhase
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
	Count int64 `json:"count"`
}

type DeleteManyEventsResult struct {
	Count int64 `json:"count"`
}

// EventsClient is the specialized client for managing Events with the BrigNext
// API.
type EventsClient interface {
	// Create creates a new Event.
	Create(context.Context, sdk.Event) (EventList, error)
	// List returns an EventList, with its Items (Events) ordered by age, newest
	// first. Criteria for which Events should be retrieved can be specified using
	// the EventListOptions parameter.
	List(context.Context, EventsSelector, meta.ListOptions) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (sdk.Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(context.Context, EventsSelector) (CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(context.Context, EventsSelector) (DeleteManyEventsResult, error)

	Workers() WorkersClient

	Logs() LogsClient
}

type eventsClient struct {
	*baseClient
	workersClient WorkersClient
	logsClient    LogsClient
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
		workersClient: NewWorkersClient(apiAddress, apiToken, allowInsecure),
		logsClient:    NewLogsClient(apiAddress, apiToken, allowInsecure),
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
	selector EventsSelector,
	opts meta.ListOptions,
) (EventList, error) {
	queryParams := map[string]string{}
	if selector.ProjectID != "" {
		queryParams["projectID"] = selector.ProjectID
	}
	if len(selector.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(selector.WorkerPhases))
		for i, workerPhase := range selector.WorkerPhases {
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
			queryParams: e.appendListQueryParams(queryParams, opts),
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
	opts EventsSelector,
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
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	queryParams := map[string]string{}
	if selector.ProjectID != "" {
		queryParams["projectID"] = selector.ProjectID
	}
	if len(selector.WorkerPhases) > 0 {
		workerPhaseStrs := make([]string, len(selector.WorkerPhases))
		for i, workerPhase := range selector.WorkerPhases {
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

func (e *eventsClient) Workers() WorkersClient {
	return e.workersClient
}

func (e *eventsClient) Logs() LogsClient {
	return e.logsClient
}
