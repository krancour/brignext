package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type ServiceAccountsClient interface {
	Create(context.Context, ServiceAccount) (Token, error)
	List(context.Context) (ServiceAccountList, error)
	Get(context.Context, string) (ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (Token, error)
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
	serviceAccount ServiceAccount,
) (Token, error) {
	token := Token{}
	return token, s.executeAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/service-accounts",
			authHeaders: s.bearerTokenAuthHeaders(),
			reqBodyObj:  serviceAccount,
			successCode: http.StatusCreated,
			respObj:     &token,
		},
	)
}

func (s *serviceAccountsClient) List(
	context.Context,
) (ServiceAccountList, error) {
	serviceAccountList := ServiceAccountList{}
	return serviceAccountList, s.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/service-accounts",
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &serviceAccountList,
		},
	)
}

func (s *serviceAccountsClient) Get(
	_ context.Context,
	id string,
) (ServiceAccount, error) {
	serviceAccount := ServiceAccount{}
	return serviceAccount, s.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/service-accounts/%s", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &serviceAccount,
		},
	)
}

func (s *serviceAccountsClient) Lock(_ context.Context, id string) error {
	return s.executeAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (s *serviceAccountsClient) Unlock(
	_ context.Context,
	id string,
) (Token, error) {
	token := Token{}
	return token, s.executeAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &token,
		},
	)
}
