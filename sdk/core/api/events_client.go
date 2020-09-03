package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/krancour/brignext/v2/sdk/core"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// EventsClient is the specialized client for managing Events with the BrigNext
// API.
type EventsClient interface {
	// Create creates a new Event.
	Create(context.Context, core.Event) (EventList, error)
	// List returns an EventList, with its Items (Events) ordered by age, newest
	// first. Criteria for which Events should be retrieved can be specified using
	// the EventListOptions parameter.
	List(context.Context, EventsSelector, meta.ListOptions) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (core.Event, error)
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
	*apimachinery.BaseClient
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
		workersClient: NewWorkersClient(apiAddress, apiToken, allowInsecure),
		logsClient:    NewLogsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (e *eventsClient) Create(
	_ context.Context,
	event core.Event,
) (EventList, error) {
	events := EventList{}
	return events, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			ReqBodyObj:  event,
			SuccessCode: http.StatusCreated,
			RespObj:     &events,
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
	return events, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: e.AppendListQueryParams(queryParams, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &events,
		},
	)
}

func (e *eventsClient) Get(
	_ context.Context,
	id string,
) (core.Event, error) {
	event := core.Event{}
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
	return result, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/events/cancellations",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &result,
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
	return result, e.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/events",
			AuthHeaders: e.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
			RespObj:     &result,
		},
	)
}

func (e *eventsClient) Workers() WorkersClient {
	return e.workersClient
}

func (e *eventsClient) Logs() LogsClient {
	return e.logsClient
}
