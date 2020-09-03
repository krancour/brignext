package api

import (
	authx "github.com/krancour/brignext/v2/sdk/authx/api"
	core "github.com/krancour/brignext/v2/sdk/core/api"
	system "github.com/krancour/brignext/v2/sdk/system/api"
)

// Client is the general interface for the BrigNext API. It does little more
// than expose functions for obtaining more specialized clients for different
// areas of concern, like User management or Project management.
type Client interface {
	Authx() authx.Client
	Core() core.Client
	System() system.Client
}

type client struct {
	authxClient  authx.Client
	coreClient   core.Client
	systemClient system.Client
}

// NewClient returns a BrigNext client.
func NewClient(apiAddress, apiToken string, allowInsecure bool) Client {
	return &client{
		authxClient:  authx.NewClient(apiAddress, apiToken, allowInsecure),
		coreClient:   core.NewClient(apiAddress, apiToken, allowInsecure),
		systemClient: system.NewClient(apiAddress, apiToken, allowInsecure),
	}
}

func (c *client) Authx() authx.Client {
	return c.authxClient
}

func (c *client) Core() core.Client {
	return c.coreClient
}

func (c *client) System() system.Client {
	return c.systemClient
}
