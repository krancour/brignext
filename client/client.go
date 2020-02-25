package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

type Client interface {
	GetUsers(context.Context) ([]brignext.User, error)
	GetUser(context.Context, string) (brignext.User, error)
	LockUser(context.Context, string) error
	UnlockUser(context.Context, string) error

	CreateRootSession(ctx context.Context, password string) (string, error)
	CreateUserSession(context.Context) (string, string, error)
	DeleteSession(context.Context) error

	CreateServiceAccount(context.Context, brignext.ServiceAccount) (string, error)
	GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error)
	GetServiceAccount(
		context.Context,
		string,
	) (brignext.ServiceAccount, error)
	LockServiceAccount(context.Context, string) error
	UnlockServiceAccount(context.Context, string) (string, error)

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) ([]brignext.Project, error)
	GetProject(context.Context, string) (brignext.Project, error)
	UpdateProject(context.Context, brignext.Project) error
	DeleteProject(context.Context, string) error

	CreateEvent(context.Context, brignext.Event) (string, error)
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	UpdateEventWorkers(
		ctx context.Context,
		id string,
		workers map[string]brignext.Worker,
	) error
	UpdateEventStatus(
		ctx context.Context,
		id string,
		status brignext.EventStatus,
	) error
	UpdateEventWorkerStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		status brignext.WorkerStatus,
	) error
	GetEvent(context.Context, string) (brignext.Event, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (int64, error)
}

type client struct {
	apiAddress string
	apiToken   string
	httpClient *http.Client
}

func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
		apiAddress: apiAddress,
		apiToken:   apiToken,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: allowInsecure,
				},
			},
		},
	}
}

func (c *client) GetUsers(context.Context) ([]brignext.User, error) {
	req, err := c.buildRequest(http.MethodGet, "v2/users", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	users := []brignext.User{}
	if err := json.Unmarshal(respBodyBytes, &users); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return users, nil
}

func (c *client) GetUser(_ context.Context, id string) (brignext.User, error) {
	user := brignext.User{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/users/%s", id),
		nil,
	)
	if err != nil {
		return user, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return user, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return user, &brignext.ErrUserNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return user, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return user, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &user); err != nil {
		return user, errors.Wrap(err, "error unmarshaling response body")
	}

	return user, nil
}

func (c *client) LockUser(_ context.Context, id string) error {
	req, err := c.buildRequest(
		http.MethodPost,
		fmt.Sprintf("v2/users/%s/lock", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrUserNotFound{id}
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UnlockUser(_ context.Context, id string) error {
	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/users/%s/lock", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrUserNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateRootSession(_ context.Context, password string) (string, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/sessions", c.apiAddress),
		nil,
	)
	if err != nil {
		return "", errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("root", "true")
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth("root", password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return "", errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Token, nil
}

func (c *client) CreateUserSession(context.Context) (string, string, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v2/sessions", c.apiAddress),
		nil,
	)
	if err != nil {
		return "", "", errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", "", errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token   string `json:"token"`
		AuthURL string `json:"authURL"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return "", "", errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.AuthURL, respStruct.Token, nil
}

func (c *client) DeleteSession(context.Context) error {
	req, err := c.buildRequest(http.MethodDelete, "v2/session", nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateServiceAccount(_ context.Context, serviceAccount brignext.ServiceAccount) (string, error) {
	serviceAccountBytes, err := json.Marshal(serviceAccount)
	if err != nil {
		return "", errors.Wrap(err, "error marshaling service account")
	}

	req, err := c.buildRequest(
		http.MethodPost,
		"v2/service-accounts",
		serviceAccountBytes,
	)
	if err != nil {
		return "", errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", &brignext.ErrServiceAccountIDConflict{serviceAccount.ID}
	}
	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return "", errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Token, nil
}

func (c *client) GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error) {
	req, err := c.buildRequest(http.MethodGet, "v2/service-accounts", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	serviceAccounts := []brignext.ServiceAccount{}
	if err := json.Unmarshal(respBodyBytes, &serviceAccounts); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return serviceAccounts, nil
}

func (c *client) GetServiceAccount(
	_ context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/service-accounts/%s", id),
		nil,
	)
	if err != nil {
		return serviceAccount, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return serviceAccount, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return serviceAccount, &brignext.ErrServiceAccountNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return serviceAccount, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return serviceAccount, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &serviceAccount); err != nil {
		return serviceAccount, errors.Wrap(err, "error unmarshaling response body")
	}

	return serviceAccount, nil
}

func (c *client) LockServiceAccount(_ context.Context, id string) error {
	req, err := c.buildRequest(
		http.MethodPost,
		fmt.Sprintf("v2/service-accounts/%s/lock", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrServiceAccountNotFound{id}
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UnlockServiceAccount(_ context.Context, id string) (string, error) {
	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/service-accounts/%s/lock", id),
		nil,
	)
	if err != nil {
		return "", errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", &brignext.ErrServiceAccountNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return "", errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Token, nil
}

func (c *client) CreateProject(_ context.Context, project brignext.Project) error {
	projectBytes, err := json.Marshal(project)
	if err != nil {
		return errors.Wrap(err, "error marshaling project")
	}

	req, err := c.buildRequest(http.MethodPost, "v2/projects", projectBytes)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return &brignext.ErrProjectIDConflict{project.ID}
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) GetProjects(context.Context) ([]brignext.Project, error) {
	req, err := c.buildRequest(http.MethodGet, "v2/projects", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	projects := []brignext.Project{}
	if err := json.Unmarshal(respBodyBytes, &projects); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return projects, nil
}

func (c *client) GetProject(_ context.Context, id string) (brignext.Project, error) {
	project := brignext.Project{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/projects/%s", id),
		nil,
	)
	if err != nil {
		return project, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return project, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return project, &brignext.ErrProjectNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return project, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return project, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &project); err != nil {
		return project, errors.Wrap(err, "error unmarshaling response body")
	}

	return project, nil
}

func (c *client) UpdateProject(_ context.Context, project brignext.Project) error {
	projectBytes, err := json.Marshal(project)
	if err != nil {
		return errors.Wrap(err, "error marshaling project")
	}

	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/projects/%s", project.ID),
		projectBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrProjectNotFound{project.ID}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) DeleteProject(_ context.Context, id string) error {
	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrProjectNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateEvent(_ context.Context, event brignext.Event) (string, error) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return "", errors.Wrap(err, "error marshaling event")
	}

	req, err := c.buildRequest(http.MethodPost, "v2/events", eventBytes)
	if err != nil {
		return "", errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", &brignext.ErrProjectNotFound{event.ProjectID}
	}
	if resp.StatusCode != http.StatusCreated {
		return "", errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return "", errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.ID, nil
}

func (c *client) GetEvents(context.Context) ([]brignext.Event, error) {
	req, err := c.buildRequest(http.MethodGet, "v2/events", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	events := []brignext.Event{}
	if err := json.Unmarshal(respBodyBytes, &events); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return events, nil
}

func (c *client) GetEventsByProject(_ context.Context, projectID string) ([]brignext.Event, error) {
	req, err := c.buildRequest(http.MethodGet, "v2/events", nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}
	if projectID != "" {
		q := req.URL.Query()
		q.Set("projectID", projectID)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &brignext.ErrProjectNotFound{projectID}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	events := []brignext.Event{}
	if err := json.Unmarshal(respBodyBytes, &events); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return events, nil
}

func (c *client) GetEvent(ctx context.Context, id string) (brignext.Event, error) {
	event := brignext.Event{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s", id),
		nil,
	)
	if err != nil {
		return event, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return event, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return event, &brignext.ErrEventNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return event, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return event, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &event); err != nil {
		return event, errors.Wrap(err, "error unmarshaling response body")
	}

	return event, nil
}

func (c *client) UpdateEventWorkers(
	ctx context.Context,
	id string,
	workers map[string]brignext.Worker,
) error {
	workersBytes, err := json.Marshal(workers)
	if err != nil {
		return errors.Wrap(err, "error marshaling workers")
	}

	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/events/%s/workers", id),
		workersBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrEventNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UpdateEventWorkerStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	status brignext.WorkerStatus,
) error {
	statusBytes, err := json.Marshal(
		struct {
			Status brignext.WorkerStatus `json:"status"`
		}{
			Status: status,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/events/%s/workers/%s/status", eventID, workerName),
		statusBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UpdateEventStatus(
	ctx context.Context,
	id string,
	status brignext.EventStatus,
) error {
	statusBytes, err := json.Marshal(
		struct {
			Status brignext.EventStatus `json:"status"`
		}{
			Status: status,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/events/%s/status", id),
		statusBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &brignext.ErrEventNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) DeleteEvent(
	ctx context.Context,
	id string,
	deleteAccepted bool,
	deleteProcessing bool,
) (bool, error) {
	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/events/%s", id),
		nil,
	)
	if err != nil {
		return false, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if deleteAccepted {
		q.Set("deleteAccepted", "true")
	}
	if deleteProcessing {
		q.Set("deleteProcessing", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, &brignext.ErrEventNotFound{id}
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Deleted bool `json:"deleted"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return false, errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Deleted, nil
}

func (c *client) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
	deleteAccepted bool,
	deleteProcessing bool,
) (int64, error) {
	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s/events", projectID),
		nil,
	)
	if err != nil {
		return 0, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if deleteAccepted {
		q.Set("deleteAccepted", "true")
	}
	if deleteProcessing {
		q.Set("deleteProcessing", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, &brignext.ErrProjectNotFound{projectID}
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Deleted int64 `json:"deleted"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return 0, errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Deleted, nil
}

func (c *client) buildRequest(
	method string,
	path string,
	body []byte,
) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", c.apiAddress, path),
		bodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating request %s %s", method, path)
	}

	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", c.apiToken),
	)

	return req, nil
}
