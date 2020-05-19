package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type JobsClient interface {
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
	GetLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
	GetInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (LogEntryList, error)
	StreamInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
}

type jobsClient struct {
	*baseClient
}

func NewJobsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) JobsClient {
	return &jobsClient{
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

func (j *jobsClient) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	return j.doAPIRequest(
		apiRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/status",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			reqBodyObj:  status,
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
}

func (j *jobsClient) GetLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	err := j.doAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &logEntryList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	return logEntryList, err
}

func (j *jobsClient) StreamLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := j.doAPIRequest2(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
			},
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go j.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (j *jobsClient) GetInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (LogEntryList, error) {
	logEntryList := LogEntryList{}
	err := j.doAPIRequest(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"init": "true",
			},
			successCode: http.StatusOK,
			respObj:     &logEntryList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	return logEntryList, err
}

func (j *jobsClient) StreamInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	resp, err := j.doAPIRequest2(
		apiRequest{
			method: http.MethodGet,
			path: fmt.Sprintf(
				"v2/events/%s/worker/jobs/%s/logs",
				eventID,
				jobName,
			),
			authHeaders: j.bearerTokenAuthHeaders(),
			queryParams: map[string]string{
				"stream": "true",
				"init":   "true",
			},
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrJobNotFound{},
			},
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go j.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}
