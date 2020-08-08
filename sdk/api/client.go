package api

// Client is the general interface for the BrigNext API. It does little more
// than expose functions for obtaining more specialized clients for different
// areas of concern, like User management or Project management.
type Client interface {
	// Events returns a specialized client for Event management.
	Events() EventsClient
	// Projects returns a specialized client for Project management.
	Projects() ProjectsClient
	// ServiceAccounts returns a specialized client for ServiceAccount management.
	ServiceAccounts() ServiceAccountsClient
	// Sessions returns a specialized client for Session management.
	Sessions() SessionsClient
	// Users returns a specialized client for User management.
	Users() UsersClient
}

type client struct {
	// eventsClient is a specialized client for Event management.
	eventsClient EventsClient
	// projectsClient is a specialized client for Project management.
	projectsClient ProjectsClient
	// serviceAccountsClient is a specialized client for ServiceAccount
	// management.
	serviceAccountsClient ServiceAccountsClient
	// sessionsClient is a specialized client for Session management.
	sessionsClient SessionsClient
	// usersClient is a specialized client for User managament.
	usersClient UsersClient
}

// NewClient returns a BrigNext client.
func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
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
