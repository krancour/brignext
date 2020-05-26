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

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
	GetWorkerLogs(ctx context.Context, eventID string) (LogEntryList, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
	) (<-chan LogEntry, <-chan error, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (LogEntryList, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan LogEntry, <-chan error, error)

	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
	GetJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (LogEntryList, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (LogEntryList, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
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
	return eventRefList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  event,
			successCode: http.StatusCreated,
			respObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) List(context.Context) (EventList, error) {
	eventList := EventList{}
	return eventList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &eventList,
		},
	)
}

func (e *eventsClient) ListByProject(
	_ context.Context,
	projectID string,
) (EventList, error) {
	eventList := EventList{}
	return eventList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/events",
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"projectID": projectID,
			},
			successCode: http.StatusOK,
			respObj:     &eventList,
		},
	)
}

func (e *eventsClient) Get(ctx context.Context, id string) (Event, error) {
	event := Event{}
	return event, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &event,
		},
	)
}

func (e *eventsClient) Cancel(
	ctx context.Context,
	id string,
	cancelRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/cancel", id),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"cancelRunning": strconv.FormatBool(cancelRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) CancelByProject(
	ctx context.Context,
	projectID string,
	cancelRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s/events/cancel", projectID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"cancelRunning": strconv.FormatBool(cancelRunning),
			},
			successCode: http.StatusOK,
			respObj:     &eventRefList,
		},
	)
}

func (e *eventsClient) Delete(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeAPIRequest(
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
		},
	)
}

func (e *eventsClient) DeleteByProject(
	ctx context.Context,
	projectID string,
	deletePending bool,
	deleteRunning bool,
) (EventReferenceList, error) {
	eventRefList := EventReferenceList{}
	return eventRefList, e.executeAPIRequest(
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
		},
	)
}

func (e *eventsClient) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	return e.executeAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/events/%s/worker/status", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
		},
	)
}

func (e *eventsClient) GetWorkerLogs(
	ctx context.Context,
	eventID string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	return logEntryList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &logEntryList,
		},
	)
}

func (e *eventsClient) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := e.submitAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (e *eventsClient) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	return logEntryList, e.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"init": "true",
			},
			successCode: http.StatusOK,
			respObj:     &logEntryList,
		},
	)
}

func (e *eventsClient) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := e.submitAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/events/%s/worker/logs", eventID),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
				"init":   "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (e *eventsClient) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	return e.executeAPIRequest(
		apiRequest{
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

func (e *eventsClient) GetJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	return logEntryList, e.executeAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &logEntryList,
		},
	)
}

func (e *eventsClient) StreamJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := e.submitAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (e *eventsClient) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	return logEntryList, e.executeAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"init": "true",
			},
			successCode: http.StatusOK,
			respObj:     &logEntryList,
		},
	)
}

func (e *eventsClient) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := e.submitAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: e.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
				"init":   "true",
			},
			successCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go e.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}
