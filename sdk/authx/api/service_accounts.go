package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/authx"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// ServiceAccountsSelector represents useful filter criteria when selecting
// multiple ServiceAccounts for API group operations like list. It currently has
// no fields, but exists for future expansion.
type ServiceAccountsSelector struct{}

// ServiceAccountList is an ordered and pageable list of ServiceAccounts.
type ServiceAccountList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of ServiceAccounts.
	Items []authx.ServiceAccount `json:"items,omitempty"`
}

// MarshalJSON amends ServiceAccountList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (s ServiceAccountList) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountList",
			},
			Alias: (Alias)(s),
		},
	)
}

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
