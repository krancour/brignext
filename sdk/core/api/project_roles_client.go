package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	authx "github.com/krancour/brignext/v2/sdk/authx/api"
	"github.com/krancour/brignext/v2/sdk/internal/restmachinery"
)

// ProjectRolesClient is the specialized client for managing Project Roles with
// the BrigNext API.
type ProjectRolesClient interface {
	GrantToUser(
		ctx context.Context,
		projectID string,
		userID string,
		roleName string,
	) error
	RevokeFromUser(
		ctx context.Context,
		projectID string,
		userID string,
		roleName string,
	) error

	GrantToServiceAccount(
		ctx context.Context,
		projectID string,
		serviceAccountID string,
		roleName string,
	) error
	RevokeFromServiceAccount(
		ctx context.Context,
		projectID string,
		serviceAccountID string,
		roleName string,
	) error
}

type projectRolesClient struct {
	*restmachinery.BaseClient
}

// NewProjectRolesClient returns a specialized client for managing Project
// Roles.
func NewProjectRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectRolesClient {
	return &projectRolesClient{
		BaseClient: &restmachinery.BaseClient{
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
	}
}

func (p *projectRolesClient) GrantToUser(
	ctx context.Context,
	projectID string,
	userID string,
	roleName string,
) error {
	return p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodPost,
			Path: fmt.Sprintf(
				"v2/projects/%s/user-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj: authx.UserRoleAssignment{
				UserID: userID,
				Role:   roleName,
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRolesClient) RevokeFromUser(
	ctx context.Context,
	projectID string,
	userID string,
	roleName string,
) error {
	queryParams := map[string]string{
		"userID": userID,
		"role":   roleName,
	}
	return p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/user-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRolesClient) GrantToServiceAccount(
	ctx context.Context,
	projectID string,
	serviceAccountID string,
	roleName string,
) error {
	return p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodPost,
			Path: fmt.Sprintf(
				"v2/projects/%s/service-account-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj: authx.ServiceAccountRoleAssignment{
				ServiceAccountID: serviceAccountID,
				Role:             roleName,
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRolesClient) RevokeFromServiceAccount(
	ctx context.Context,
	projectID string,
	serviceAccountID string,
	roleName string,
) error {
	queryParams := map[string]string{
		"serviceAccountID": serviceAccountID,
		"role":             roleName,
	}
	return p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/service-account-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
