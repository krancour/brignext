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
	err := s.doAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/service-accounts",
			authHeaders: s.bearerTokenAuthHeaders(),
			reqBodyObj:  serviceAccount,
			successCode: http.StatusCreated,
			respObj:     &token,
		},
	)
	return token, err
}

func (s *serviceAccountsClient) List(
	context.Context,
) (ServiceAccountList, error) {
	serviceAccountList := ServiceAccountList{}
	err := s.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/service-accounts",
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &serviceAccountList,
		},
	)
	return serviceAccountList, err
}

func (s *serviceAccountsClient) Get(
	_ context.Context,
	id string,
) (ServiceAccount, error) {
	serviceAccount := ServiceAccount{}
	err := s.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/service-accounts/%s", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &serviceAccount,
		},
	)
	return serviceAccount, err
}

func (s *serviceAccountsClient) Lock(_ context.Context, id string) error {
	return s.doAPIRequest(
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
	err := s.doAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &token,
		},
	)
	return token, err
}
