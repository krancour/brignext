package api

import (
	"crypto/tls"
	"net/http"

	"github.com/krancour/brignext/v2/api/events"
	"github.com/krancour/brignext/v2/api/projects"
	"github.com/krancour/brignext/v2/api/serviceaccounts"
	"github.com/krancour/brignext/v2/api/sessions"
	"github.com/krancour/brignext/v2/api/users"
	"github.com/krancour/brignext/v2/internal/api"
)

type Client interface {
	Events() events.Client
	Projects() projects.Client
	ServiceAccounts() serviceaccounts.Client
	Sessions() sessions.Client
	Users() users.Client
}

type client struct {
	*api.BaseClient
	eventsClient          events.Client
	projectsClient        projects.Client
	serviceAccountsClient serviceaccounts.Client
	sessionsClient        sessions.Client
	usersClient           users.Client
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
		sessionsClient: sessions.NewClient(apiAddress, apiToken, allowInsecure),
		usersClient:    users.NewClient(apiAddress, apiToken, allowInsecure),
		serviceAccountsClient: serviceaccounts.NewClient(
			apiAddress,
			apiToken,
			allowInsecure,
		),
		projectsClient: projects.NewClient(apiAddress, apiToken, allowInsecure),
		eventsClient:   events.NewClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Events() events.Client {
	return c.eventsClient
}

func (c *client) Projects() projects.Client {
	return c.projectsClient
}

func (c *client) ServiceAccounts() serviceaccounts.Client {
	return c.serviceAccountsClient
}

func (c *client) Sessions() sessions.Client {
	return c.sessionsClient
}

func (c *client) Users() users.Client {
	return c.usersClient
}
