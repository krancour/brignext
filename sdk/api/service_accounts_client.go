package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
)

type ServiceAccountsClient interface {
	Create(context.Context, brignext.ServiceAccount) (brignext.Token, error)
	List(context.Context) (brignext.ServiceAccountReferenceList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (brignext.Token, error)
}

type serviceAccountsClient struct {
	*apimachinery.BaseClient
}

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
	serviceAccount brignext.ServiceAccount,
) (brignext.Token, error) {
	token := brignext.Token{}
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
	context.Context,
) (brignext.ServiceAccountReferenceList, error) {
	serviceAccountList := brignext.ServiceAccountReferenceList{}
	return serviceAccountList, s.ExecuteRequest(
		apimachinery.OutboundRequest{
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
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
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
) (brignext.Token, error) {
	token := brignext.Token{}
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
