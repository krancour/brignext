package api

type Client interface {
	// ServiceAccounts returns a specialized client for ServiceAccount management.
	ServiceAccounts() ServiceAccountsClient
	// Sessions returns a specialized client for Session management.
	Sessions() SessionsClient
	// Users returns a specialized client for User management.
	Users() UsersClient
}

type client struct {
	// serviceAccountsClient is a specialized client for ServiceAccount
	// management.
	serviceAccountsClient ServiceAccountsClient
	// sessionsClient is a specialized client for Session management.
	sessionsClient SessionsClient
	// usersClient is a specialized client for User managament.
	usersClient UsersClient
}

func NewClient(
	apiAddress,
	apiToken string,
	allowInsecure bool,
) Client {
	return &client{
		serviceAccountsClient: NewServiceAccountsClient(
			apiAddress,
			apiToken,
			allowInsecure,
		),
		sessionsClient: NewSessionsClient(apiAddress, apiToken, allowInsecure),
		usersClient:    NewUsersClient(apiAddress, apiToken, allowInsecure),
	}
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
