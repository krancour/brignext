package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/brigadecore/brigade/v2/sdk/core"
	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
	"github.com/brigadecore/brigade/v2/sdk/meta"
)

// SecretsClient is the specialized client for managing Secrets with the
// Brigade API.
type SecretsClient interface {
	// List returns a SecretList who Items (Secrets) contain Keys only and not
	// Values (all Value fields are empty). i.e. Once a secret is set, end clients
	// are unable to retrieve values.
	List(
		ctx context.Context,
		projectID string,
		opts meta.ListOptions,
	) (SecretList, error)
	// Set sets the value of a new Secret or updates the value of an existing
	// Secret.
	Set(ctx context.Context, projectID string, secret core.Secret) error
	// Unset clears the value of an existing Secret.
	Unset(ctx context.Context, projectID string, key string) error
}

type secretsClient struct {
	*restmachinery.BaseClient
}

// NewSecretsClient returns a specialized client for managing
// Secrets.
func NewSecretsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SecretsClient {
	return &secretsClient{
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

func (s *secretsClient) List(
	ctx context.Context,
	projectID string,
	opts meta.ListOptions,
) (SecretList, error) {
	secrets := SecretList{}
	return secrets, s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			QueryParams: s.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &secrets,
		},
	)
}

func (s *secretsClient) Set(
	ctx context.Context,
	projectID string,
	secret core.Secret,
) error {
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
			),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			ReqBodyObj:  secret,
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *secretsClient) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	return s.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
