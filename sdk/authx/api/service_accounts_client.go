package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/authx"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// ServiceAccountsClient is the specialized client for managing ServiceAccounts
// with the BrigNext API.
type ServiceAccountsClient interface {
	// Create creates a new ServiceAccount.
	Create(context.Context, authx.ServiceAccount) (Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		ServiceAccountsSelector,
		meta.ListOptions,
	) (ServiceAccountList, error)
	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (authx.ServiceAccount, error)

	// Lock removes access to the API for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single ServiceAccount specified by
	// its identifier. It returns a new Token.
	Unlock(context.Context, string) (Token, error)
}

type serviceAccountsClient struct {
	*apimachinery.BaseClient
}

// NewServiceAccountsClient returns a specialized client for managing
// ServiceAccounts.
func NewServiceAccountsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ServiceAccountsClient {
	return &serviceAccountsClient{
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

func (s *serviceAccountsClient) Create(
	_ context.Context,
	serviceAccount authx.ServiceAccount,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/service-accounts",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			ReqBodyObj:  serviceAccount,
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (s *serviceAccountsClient) List(
	_ context.Context,
	_ ServiceAccountsSelector,
	opts meta.ListOptions,
) (ServiceAccountList, error) {
	serviceAccounts := ServiceAccountList{}
	return serviceAccounts, s.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/service-accounts",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			QueryParams: s.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccounts,
		},
	)
}

func (s *serviceAccountsClient) Get(
	_ context.Context,
	id string,
) (authx.ServiceAccount, error) {
	serviceAccount := authx.ServiceAccount{}
	return serviceAccount, s.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/service-accounts/%s", id),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccount,
		},
	)
}

func (s *serviceAccountsClient) Lock(_ context.Context, id string) error {
	return s.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *serviceAccountsClient) Unlock(
	_ context.Context,
	id string,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &token,
		},
	)
}
