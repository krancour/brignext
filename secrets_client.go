package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type SecretsClient interface {
	List(ctx context.Context, projectID string) (SecretList, error)
	Set(ctx context.Context, projectID string, secret Secret) error
	Unset(ctx context.Context, projectID string, secretID string) error
}

type secretsClient struct {
	*baseClient
}

func NewSecretsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SecretsClient {
	return &secretsClient{
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

func (s *secretsClient) List(
	ctx context.Context,
	projectID string,
) (SecretList, error) {
	secretList := SecretList{}
	err := s.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &secretList,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
	return secretList, err
}

func (s *secretsClient) Set(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	return s.doAPIRequest(
		apiRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.ID,
			),
			authHeaders: s.bearerTokenAuthHeaders(),
			reqBodyObj:  secret,
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
}

func (s *secretsClient) Unset(
	ctx context.Context,
	projectID string,
	secretID string,
) error {
	return s.doAPIRequest(
		apiRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secretID,
			),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			errObjs: map[int]error{
				http.StatusNotFound: &ErrProjectNotFound{},
			},
		},
	)
}
