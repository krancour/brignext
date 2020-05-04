package brignext

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Client interface {
	GetUsers(context.Context) ([]User, error)
	GetUser(context.Context, string) (User, error)
	LockUser(context.Context, string) error
	UnlockUser(context.Context, string) error

	CreateRootSession(ctx context.Context, password string) (string, error)
	CreateUserSession(context.Context) (string, string, error)
	DeleteSession(context.Context) error

	CreateServiceAccount(context.Context, ServiceAccount) (string, error)
	GetServiceAccounts(context.Context) ([]ServiceAccount, error)
	GetServiceAccount(
		context.Context,
		string,
	) (ServiceAccount, error)
	LockServiceAccount(context.Context, string) error
	UnlockServiceAccount(context.Context, string) (string, error)

	CreateProject(context.Context, Project) error
	GetProjects(context.Context) ([]Project, error)
	GetProject(context.Context, string) (Project, error)
	UpdateProject(context.Context, Project) error
	DeleteProject(context.Context, string) error

	GetSecrets(
		ctx context.Context,
		projectID string,
		workerName string,
	) (map[string]string, error)
	SetSecrets(
		ctx context.Context,
		projectID string,
		workerName string,
		secrets map[string]string,
	) error
	UnsetSecrets(
		ctx context.Context,
		projectID string,
		workerName string,
		keys []string,
	) error

	CreateEvent(context.Context, Event) (string, error)
	GetEvents(context.Context) ([]Event, error)
	GetEventsByProject(context.Context, string) ([]Event, error)
	GetEvent(context.Context, string) (Event, error)
	CancelEvent(
		ctx context.Context,
		id string,
		cancelProcessing bool,
	) (bool, error)
	CancelEventsByProject(
		ctx context.Context,
		projectID string,
		cancelProcessing bool,
	) (int64, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteProcessing bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deletePending bool,
		deleteProcessing bool,
	) (int64, error)

	GetWorker(
		ctx context.Context,
		eventID string,
		workerName string,
	) (Worker, error)
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		status WorkerStatus,
	) error
	CancelWorker(
		ctx context.Context,
		eventID string,
		workerName string,
		cancelRunning bool,
	) (bool, error)
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]LogEntry, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan LogEntry, <-chan error, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan LogEntry, <-chan error, error)

	GetJob(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (Job, error)
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
		status JobStatus,
	) error
	GetJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]LogEntry, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (<-chan LogEntry, <-chan error, error)
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

func (c *client) GetUsers(context.Context) ([]User, error) {
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

	users := []User{}
	if err := json.Unmarshal(respBodyBytes, &users); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return users, nil
}

func (c *client) GetUser(_ context.Context, id string) (User, error) {
	user := User{}
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
		return user, &ErrUserNotFound{
			ID: id,
		}
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
		return &ErrUserNotFound{
			ID: id,
		}
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
		return &ErrUserNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateRootSession(
	_ context.Context,
	password string,
) (string, error) {
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

func (c *client) CreateServiceAccount(
	_ context.Context,
	serviceAccount ServiceAccount,
) (string, error) {
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
		return "", &ErrServiceAccountIDConflict{
			ID: serviceAccount.ID,
		}
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

func (c *client) GetServiceAccounts(
	context.Context,
) ([]ServiceAccount, error) {
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

	serviceAccounts := []ServiceAccount{}
	if err := json.Unmarshal(respBodyBytes, &serviceAccounts); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return serviceAccounts, nil
}

func (c *client) GetServiceAccount(
	_ context.Context,
	id string,
) (ServiceAccount, error) {
	serviceAccount := ServiceAccount{}
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
		return serviceAccount, &ErrServiceAccountNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return serviceAccount, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
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
		return &ErrServiceAccountNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UnlockServiceAccount(
	_ context.Context,
	id string,
) (string, error) {
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
		return "", &ErrServiceAccountNotFound{
			ID: id,
		}
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

func (c *client) CreateProject(
	_ context.Context,
	project Project,
) error {
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
		return &ErrProjectIDConflict{
			ID: project.ID,
		}
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) GetProjects(context.Context) ([]Project, error) {
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

	projects := []Project{}
	if err := json.Unmarshal(respBodyBytes, &projects); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return projects, nil
}

func (c *client) GetProject(_ context.Context, id string) (Project, error) {
	project := Project{}
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
		return project, &ErrProjectNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return project, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
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

func (c *client) UpdateProject(
	_ context.Context,
	project Project,
) error {
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
		return &ErrProjectNotFound{
			ID: project.ID,
		}
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
		return &ErrProjectNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) GetSecrets(
	ctx context.Context,
	projectID string,
	workerName string,
) (map[string]string, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/projects/%s/workers/%s/secrets", projectID, workerName),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &ErrWorkerNotFound{
			ProjectID:  projectID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	secrets := map[string]string{}
	if err := json.Unmarshal(respBodyBytes, &secrets); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return secrets, nil
}

func (c *client) SetSecrets(
	ctx context.Context,
	projectID string,
	workerName string,
	secrets map[string]string,
) error {
	secretsBytes, err := json.Marshal(secrets)
	if err != nil {
		return errors.Wrap(err, "error marshaling secrets")
	}

	req, err := c.buildRequest(
		http.MethodPost,
		fmt.Sprintf("v2/projects/%s/workers/%s/secrets", projectID, workerName),
		secretsBytes,
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
		return &ErrProjectNotFound{
			ID: projectID,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) UnsetSecrets(
	ctx context.Context,
	projectID string,
	workerName string,
	keys []string,
) error {
	keysStruct := struct {
		Keys []string `json:"keys"`
	}{
		Keys: keys,
	}
	keysBytes, err := json.Marshal(keysStruct)
	if err != nil {
		return errors.Wrap(err, "error marshaling keys")
	}

	req, err := c.buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/projects/%s/workers/%s/secrets", projectID, workerName),
		keysBytes,
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
		return &ErrProjectNotFound{
			ID: projectID,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateEvent(_ context.Context, event Event) (string, error) {
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
		return "", &ErrProjectNotFound{
			ID: event.ProjectID,
		}
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

func (c *client) GetEvents(context.Context) ([]Event, error) {
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

	events := []Event{}
	if err := json.Unmarshal(respBodyBytes, &events); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return events, nil
}

func (c *client) GetEventsByProject(
	_ context.Context,
	projectID string,
) ([]Event, error) {
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
		return nil, &ErrProjectNotFound{
			ID: projectID,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	events := []Event{}
	if err := json.Unmarshal(respBodyBytes, &events); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return events, nil
}

func (c *client) GetEvent(ctx context.Context, id string) (Event, error) {
	event := Event{}
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
		return event, &ErrEventNotFound{
			ID: id,
		}
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

func (c *client) CancelEvent(
	ctx context.Context,
	id string,
	cancelProcessing bool,
) (bool, error) {
	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/events/%s/cancel", id),
		nil,
	)
	if err != nil {
		return false, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if cancelProcessing {
		q.Set("cancelProcessing", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, &ErrEventNotFound{
			ID: id,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Canceled bool `json:"canceled"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return false, errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Canceled, nil
}

func (c *client) CancelEventsByProject(
	ctx context.Context,
	projectID string,
	cancelProcessing bool,
) (int64, error) {
	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/projects/%s/events/cancel", projectID),
		nil,
	)
	if err != nil {
		return 0, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if cancelProcessing {
		q.Set("cancelProcessing", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, &ErrProjectNotFound{
			ID: projectID,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Canceled int64 `json:"canceled"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return 0, errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Canceled, nil
}

func (c *client) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
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
	if deletePending {
		q.Set("deletePending", "true")
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
		return false, &ErrEventNotFound{
			ID: id,
		}
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
	deletePending bool,
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
	if deletePending {
		q.Set("deletePending", "true")
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
		return 0, &ErrProjectNotFound{
			ID: projectID,
		}
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

func (c *client) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	status WorkerStatus,
) error {
	statusBytes, err := json.Marshal(
		WorkerStatus{
			Started: status.Started,
			Ended:   status.Ended,
			Phase:   status.Phase,
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
		return &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) GetWorker(
	ctx context.Context,
	eventID string,
	workerName string,
) (Worker, error) {
	worker := Worker{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/workers/%s", eventID, workerName),
		nil,
	)
	if err != nil {
		return worker, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return worker, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return worker, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return worker, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return worker, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &worker); err != nil {
		return worker, errors.Wrap(err, "error unmarshaling response body")
	}

	return worker, nil
}

func (c *client) CancelWorker(
	ctx context.Context,
	eventID string,
	workerName string,
	cancelRunning bool,
) (bool, error) {
	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf("v2/events/%s/workers/%s/cancel", eventID, workerName),
		nil,
	)
	if err != nil {
		return false, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if cancelRunning {
		q.Set("cancelRunning", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Canceled bool `json:"canceled"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return false, errors.Wrap(err, "error unmarshaling response body")
	}

	return respStruct.Canceled, nil
}

func (c *client) GetWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]LogEntry, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/workers/%s/logs", eventID, workerName),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	logEntries := []LogEntry{}
	if err := json.Unmarshal(respBodyBytes, &logEntries); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return logEntries, nil
}

func (c *client) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan LogEntry, <-chan error, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/workers/%s/logs", eventID, workerName),
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("stream", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error invoking API")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go c.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (c *client) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]LogEntry, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/workers/%s/logs", eventID, workerName),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("init", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	logEntries := []LogEntry{}
	if err := json.Unmarshal(respBodyBytes, &logEntries); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return logEntries, nil
}

func (c *client) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan LogEntry, <-chan error, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/workers/%s/logs", eventID, workerName),
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("init", "true")
	q.Set("stream", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error invoking API")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go c.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (c *client) GetJob(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (Job, error) {
	job := Job{}
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s",
			eventID,
			workerName,
			jobName,
		),
		nil,
	)
	if err != nil {
		return job, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return job, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return job, &ErrJobNotFound{
			EventID:    eventID,
			WorkerName: workerName,
			JobName:    jobName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return job, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return job, errors.Wrap(err, "error reading response body")
	}

	if err := json.Unmarshal(respBodyBytes, &job); err != nil {
		return job, errors.Wrap(err, "error unmarshaling response body")
	}

	return job, nil
}

func (c *client) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
	status JobStatus,
) error {
	statusBytes, err := json.Marshal(
		JobStatus{
			Started: status.Started,
			Ended:   status.Ended,
			Phase:   status.Phase,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error marshaling status")
	}

	req, err := c.buildRequest(
		http.MethodPut,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s/status",
			eventID,
			workerName,
			jobName,
		),
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
		return &ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	return nil
}

func (c *client) GetJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]LogEntry, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s/logs",
			eventID,
			workerName,
			jobName,
		),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &ErrJobNotFound{
			EventID:    eventID,
			WorkerName: workerName,
			JobName:    jobName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	logEntries := []LogEntry{}
	if err := json.Unmarshal(respBodyBytes, &logEntries); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return logEntries, nil
}

func (c *client) StreamJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s/logs",
			eventID,
			workerName,
			jobName,
		),
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("stream", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error invoking API")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, &ErrJobNotFound{
			EventID:    eventID,
			WorkerName: workerName,
			JobName:    jobName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go c.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (c *client) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]LogEntry, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s/logs",
			eventID,
			workerName,
			jobName,
		),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("init", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &ErrJobNotFound{
			EventID:    eventID,
			WorkerName: workerName,
			JobName:    jobName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	logEntries := []LogEntry{}
	if err := json.Unmarshal(respBodyBytes, &logEntries); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling response body")
	}

	return logEntries, nil
}

func (c *client) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan LogEntry, <-chan error, error) {
	req, err := c.buildRequest(
		http.MethodGet,
		fmt.Sprintf(
			"v2/events/%s/workers/%s/jobs/%s/logs",
			eventID,
			workerName,
			jobName,
		),
		nil,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	q.Set("stream", "true")
	q.Set("init", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error invoking API")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, &ErrJobNotFound{
			EventID:    eventID,
			WorkerName: workerName,
			JobName:    jobName,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.Errorf(
			"received %d from API server",
			resp.StatusCode,
		)
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go c.receiveLogStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

func (c *client) receiveLogStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- LogEntry,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := LogEntry{}
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