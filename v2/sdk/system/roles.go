package system

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/brigadecore/brigade/v2/sdk/authx"
	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
)

// RolesClient is the specialized client for managing System Roles with
// the Brigade API.
type RolesClient interface {
	// TODO: This needs a function for listing system role assignments
	Grant(context.Context, authx.RoleAssignment) error
	Revoke(context.Context, authx.RoleAssignment) error
}

type rolesClient struct {
	*restmachinery.BaseClient
}

// NewRolesClient returns a specialized client for managing System
// Roles.
func NewRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) RolesClient {
	return &rolesClient{
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

func (r *rolesClient) Grant(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	return r.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/system/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (r *rolesClient) Revoke(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(roleAssignment.Role),
		"principalType": string(roleAssignment.PrincipalType),
		"principalID":   roleAssignment.PrincipalID,
	}
	return r.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/role-assignments",
			AuthHeaders: r.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
