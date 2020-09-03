package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	Type string `json:"type"`
	// Name is the name of a Role and has domain-specific meaning.
	Name string `json:"name"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope"`
}

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
	*baseClient
}

// NewProjectRolesClient returns a specialized client for managing Project
// Roles.
func NewProjectRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectRolesClient {
	return &projectRolesClient{
		baseClient: &baseClient{
			apiAddress: apiAddress,
			apiToken:   apiToken,
			httpClient: &http.Client{
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
	return p.executeRequest(
		outboundRequest{
			method: http.MethodPost,
			path: fmt.Sprintf(
				"v2/projects/%s/user-role-assignments",
				projectID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj: UserRoleAssignment{
				UserID: userID,
				Role:   roleName,
			},
			successCode: http.StatusOK,
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
	return p.executeRequest(
		outboundRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/user-role-assignments",
				projectID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
		},
	)
}

func (p *projectRolesClient) GrantToServiceAccount(
	ctx context.Context,
	projectID string,
	serviceAccountID string,
	roleName string,
) error {
	return p.executeRequest(
		outboundRequest{
			method: http.MethodPost,
			path: fmt.Sprintf(
				"v2/projects/%s/service-account-role-assignments",
				projectID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj: ServiceAccountRoleAssignment{
				ServiceAccountID: serviceAccountID,
				Role:             roleName,
			},
			successCode: http.StatusOK,
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
	return p.executeRequest(
		outboundRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/service-account-role-assignments",
				projectID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			queryParams: queryParams,
			successCode: http.StatusOK,
		},
	)
}
