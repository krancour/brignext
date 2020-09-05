package api

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/brigadecore/brigade/v2/sdk/authx"
	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
)

// SystemRolesClient is the specialized client for managing System Roles with
// the Brigade API.
type SystemRolesClient interface {
	GrantRole(context.Context, authx.RoleAssignment) error
	RevokeRole(context.Context, authx.RoleAssignment) error
}

type systemRolesClient struct {
	*restmachinery.BaseClient
}

// NewSystemRolesClient returns a specialized client for managing System
// Roles.
func NewSystemRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SystemRolesClient {
	return &systemRolesClient{
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

func (s *systemRolesClient) GrantRole(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/system/role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *systemRolesClient) RevokeRole(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(roleAssignment.Role),
		"principalType": string(roleAssignment.PrincipalType),
		"principalID":   roleAssignment.PrincipalID,
	}
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
