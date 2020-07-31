package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
)

type ServiceAccountsClient interface {
	Create(context.Context, sdk.ServiceAccount) (sdk.Token, error)
	List(context.Context) (sdk.ServiceAccountReferenceList, error)
	Get(context.Context, string) (sdk.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (sdk.Token, error)
}

type serviceAccountsClient struct {
	*baseClient
}

func NewServiceAccountsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ServiceAccountsClient {
	return &serviceAccountsClient{
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

func (s *serviceAccountsClient) Create(
	_ context.Context,
	serviceAccount sdk.ServiceAccount,
) (sdk.Token, error) {
	token := sdk.Token{}
	return token, s.ExecuteRequest(
		OutboundRequest{
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
	context.Context,
) (sdk.ServiceAccountReferenceList, error) {
	serviceAccountList := sdk.ServiceAccountReferenceList{}
	return serviceAccountList, s.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/service-accounts",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccountList,
		},
	)
}

func (s *serviceAccountsClient) Get(
	_ context.Context,
	id string,
) (sdk.ServiceAccount, error) {
	serviceAccount := sdk.ServiceAccount{}
	return serviceAccount, s.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
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
) (sdk.Token, error) {
	token := sdk.Token{}
	return token, s.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &token,
		},
	)
}
