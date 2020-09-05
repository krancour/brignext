package api

import (
	"context"
	"crypto/tls"
	"net/http"

	authx "github.com/brigadecore/brigade/v2/sdk/authx/api"
	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
)

// SystemRolesClient is the specialized client for managing System Roles with
// the Brigade API.
type SystemRolesClient interface {
	GrantToUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error
	RevokeFromUser(
		ctx context.Context,
		userID string,
		roleName string,
	) error

	GrantToServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
	RevokeFromServiceAccount(
		ctx context.Context,
		serviceAccountID string,
		roleName string,
	) error
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

func (s *systemRolesClient) GrantToUser(
	ctx context.Context,
	userID string,
	roleName string,
) error {
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/system/user-role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			ReqBodyObj: authx.UserRoleAssignment{
				UserID: userID,
				Role:   roleName,
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *systemRolesClient) RevokeFromUser(
	ctx context.Context,
	userID string,
	roleName string,
) error {
	queryParams := map[string]string{
		"userID": userID,
		"role":   roleName,
	}
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/user-role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *systemRolesClient) GrantToServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roleName string,
) error {
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/system/service-account-role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			ReqBodyObj: authx.ServiceAccountRoleAssignment{
				ServiceAccountID: serviceAccountID,
				Role:             roleName,
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *systemRolesClient) RevokeFromServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roleName string,
) error {
	queryParams := map[string]string{
		"serviceAccountID": serviceAccountID,
		"role":             roleName,
	}
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/system/service-account-role-assignments",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}