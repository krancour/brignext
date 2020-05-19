package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
)

type EventsClient interface {
	Create(context.Context, Event) (EventReferenceList, error)
	List(context.Context) (EventList, error)
	ListByProject(context.Context, string) (EventList, error)
	Get(context.Context, string) (Event, error)
	Cancel(
		ctx context.Context,
		id string,
		cancelRunning bool,
	) (EventReferenceList, error)
	CancelByProject(
		ctx context.Context,
		projectID string,
		cancelRunning bool,
	) (EventReferenceList, error)
	Delete(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteRunning bool,
	) (EventReferenceList, error)
	DeleteByProject(
		ctx context.Context,
		projectID string,
		deletePending bool,
		deleteRunning bool,
	) (EventReferenceList, error)
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
	event Event,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  event,
			successCode: http.StatusCreated,
			respObj:     &eventRefList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
	return eventRefList, err
}

func (e *eventsClient) List(context.Context) (EventList, error) {
	eventList := EventList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &eventList,
		},
	)
	return eventList, err
}

func (e *eventsClient) ListByProject(
	_ context.Context,
	projectID string,
) (EventList, error) {
	eventList := EventList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"projectID": projectID,
			},
			successCode: http.StatusOK,
			respObj:     &eventList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
	return eventList, err
}

func (e *eventsClient) Get(ctx context.Context, id string) (Event, error) {
	event := Event{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &event,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
	return event, err
}

func (e *eventsClient) Cancel(
	ctx context.Context,
	id string,
	cancelRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/cancel", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"cancelRunning": strconv.FormatBool(cancelRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
	return eventRefList, err
}

func (e *eventsClient) CancelByProject(
	ctx context.Context,
	projectID string,
	cancelRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s/events/cancel", projectID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"cancelRunning": strconv.FormatBool(cancelRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
	return eventRefList, err
}

func (e *eventsClient) Delete(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/events/%s", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"deletePending": strconv.FormatBool(deletePending),
				"deleteRunning": strconv.FormatBool(deleteRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrEventNotFound{},
			},
		},
	)
	return eventRefList, err
}

func (e *eventsClient) DeleteByProject(
	ctx context.Context,
	projectID string,
	deletePending bool,
	deleteRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	err := e.doAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/projects/%s/events", projectID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"deletePending": strconv.FormatBool(deletePending),
				"deleteRunning": strconv.FormatBool(deleteRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
	return eventRefList, err
}
