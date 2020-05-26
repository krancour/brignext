package brignext

import (
	"crypto/tls"
	"net/http"
)

type Client interface {
	Events() EventsClient
	Projects() ProjectsClient
	ServiceAccounts() ServiceAccountsClient
	Sessions() SessionsClient
	Users() UsersClient
}

type client struct {
	*baseClient
	eventsClient          EventsClient
	projectsClient        ProjectsClient
	serviceAccountsClient ServiceAccountsClient
	sessionsClient        SessionsClient
	usersClient           UsersClient
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
		eventsClient:   NewEventsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Events() EventsClient {
	return c.eventsClient
}

func (c *client) Projects() ProjectsClient {
	return c.projectsClient
}

func (c *client) ServiceAccounts() ServiceAccountsClient {
	return c.serviceAccountsClient
}

func (c *client) Sessions() SessionsClient {
	return c.sessionsClient
}

func (c *client) Users() UsersClient {
	return c.usersClient
}
