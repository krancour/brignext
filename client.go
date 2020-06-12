package brignext

import (
	"crypto/tls"
	"net/http"

	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type Client interface {
	Events() EventsClient
	Projects() ProjectsClient
	ServiceAccounts() ServiceAccountsClient
	Sessions() SessionsClient
	Users() UsersClient
}

type client struct {
	*api.BaseClient
	eventsClient          EventsClient
	projectsClient        ProjectsClient
	serviceAccountsClient ServiceAccountsClient
	sessionsClient        SessionsClient
	usersClient           UsersClient
}

func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
		BaseClient: &api.BaseClient{
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
