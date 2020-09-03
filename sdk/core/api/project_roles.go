package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
)

type UserRoleAssignment struct {
	UserID string `json:"userID"`
	Role   string `json:"role"`
}

type ServiceAccountRoleAssignment struct {
	ServiceAccountID string `json:"serviceAccountID"`
	Role             string `json:"role"`
}

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
	*apimachinery.BaseClient
}

// NewProjectRolesClient returns a specialized client for managing Project
// Roles.
func NewProjectRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectRolesClient {
	return &projectRolesClient{
		BaseClient: &apimachinery.BaseClient{
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
		apimachinery.OutboundRequest{
			Method: http.MethodPost,
			Path: fmt.Sprintf(
				"v2/projects/%s/user-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj: UserRoleAssignment{
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
		apimachinery.OutboundRequest{
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
		apimachinery.OutboundRequest{
			Method: http.MethodPost,
			Path: fmt.Sprintf(
				"v2/projects/%s/service-account-role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj: ServiceAccountRoleAssignment{
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
		apimachinery.OutboundRequest{
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
