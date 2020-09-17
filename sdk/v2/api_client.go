package sdk

import (
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/system"
)

// APIClient is the general interface for the Brigade API. It does little more
// than expose functions for obtaining more specialized clients for different
// areas of concern, like User management or Project management.
type APIClient interface {
	Authx() authx.APIClient
	Core() core.APIClient
	System() system.APIClient
}

type apiClient struct {
	authxClient  authx.APIClient
	coreClient   core.APIClient
	systemClient system.APIClient
}

// NewAPIClient returns a Brigade client.
func NewAPIClient(apiAddress, apiToken string, allowInsecure bool) APIClient {
	return &apiClient{
		authxClient:  authx.NewAPIClient(apiAddress, apiToken, allowInsecure),
		coreClient:   core.NewAPIClient(apiAddress, apiToken, allowInsecure),
		systemClient: system.NewAPIClient(apiAddress, apiToken, allowInsecure),
	}
}

func (a *apiClient) Authx() authx.APIClient {
	return a.authxClient
}

func (a *apiClient) Core() core.APIClient {
	return a.coreClient
}

func (a *apiClient) System() system.APIClient {
	return a.systemClient
}
