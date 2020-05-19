package brignext

import (
	"crypto/tls"
	"net/http"
)

type Client interface {
	Sessions() SessionsClient
	Users() UsersClient
	ServiceAccounts() ServiceAccountsClient
	Projects() ProjectsClient
	Secrets() SecretsClient
	Events() EventsClient
	Workers() WorkersClient
	Jobs() JobsClient
}

type client struct {
	*baseClient
	sessionsClient        SessionsClient
	usersClient           UsersClient
	serviceAccountsClient ServiceAccountsClient
	projectsClient        ProjectsClient
	secretsClient         SecretsClient
	eventsClient          EventsClient
	workersClient         WorkersClient
	jobsClient            JobsClient
}

func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
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
		sessionsClient: NewSessionsClient(apiAddress, apiToken, allowInsecure),
		usersClient:    NewUsersClient(apiAddress, apiToken, allowInsecure),
		serviceAccountsClient: NewServiceAccountsClient(
			apiAddress,
			apiToken,
			allowInsecure,
		),
		projectsClient: NewProjectsClient(apiAddress, apiToken, allowInsecure),
		secretsClient:  NewSecretsClient(apiAddress, apiToken, allowInsecure),
		eventsClient:   NewEventsClient(apiAddress, apiToken, allowInsecure),
		workersClient:  NewWorkersClient(apiAddress, apiToken, allowInsecure),
		jobsClient:     NewJobsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Sessions() SessionsClient {
	return c.sessionsClient
}

func (c *client) Users() UsersClient {
	return c.usersClient
}

func (c *client) ServiceAccounts() ServiceAccountsClient {
	return c.serviceAccountsClient
}

func (c *client) Projects() ProjectsClient {
	return c.projectsClient
}

func (c *client) Secrets() SecretsClient {
	return c.secretsClient
}

func (c *client) Events() EventsClient {
	return c.eventsClient
}

func (c *client) Workers() WorkersClient {
	return c.workersClient
}

func (c *client) Jobs() JobsClient {
	return c.jobsClient
}
