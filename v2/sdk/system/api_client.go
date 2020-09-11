package system

type APIClient interface {
	// Roles returns a specialized client for system Role management.
	Roles() RolesClient
}

type apiClient struct {
	// rolesClient is a specialized client for system Role management.
	rolesClient RolesClient
}

func NewAPIClient(apiAddress, apiToken string, allowInsecure bool) APIClient {
	return &apiClient{
		rolesClient: NewRolesClient(apiAddress, apiToken, allowInsecure),
	}
}

func (a *apiClient) Roles() RolesClient {
	return a.rolesClient
}
