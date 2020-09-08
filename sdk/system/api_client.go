package system

type APIClient interface {
	// Roles returns a specialized client for Project Role management.
	Roles() RolesClient
}

type apiClient struct {
	// systemRolesClient is a specialized client for Event management.
	systemRolesClient RolesClient
}

func NewAPIClient(apiAddress, apiToken string, allowInsecure bool) APIClient {
	return &apiClient{
		systemRolesClient: NewRolesClient(apiAddress, apiToken, allowInsecure),
	}
}

func (a *apiClient) Roles() RolesClient {
	return a.systemRolesClient
}
