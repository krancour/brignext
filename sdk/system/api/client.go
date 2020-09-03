package api

type Client interface {
	// Roles returns a specialized client for Project Role management.
	Roles() SystemRolesClient
}

type client struct {
	// systemRolesClient is a specialized client for Event management.
	systemRolesClient SystemRolesClient
}

func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
		systemRolesClient: NewSystemRolesClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Roles() SystemRolesClient {
	return c.systemRolesClient
}
