package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []sdk.Secret `json:"items,omitempty"`
}

// MarshalJSON amends SecretList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (s SecretList) MarshalJSON() ([]byte, error) {
	type Alias SecretList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretList",
			},
			Alias: (Alias)(s),
		},
	)
}

// SecretsClient is the specialized client for managing Secrets with the
// BrigNext API.
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
	Set(ctx context.Context, projectID string, secret sdk.Secret) error
	// Unset clears the value of an existing Secret.
	Unset(ctx context.Context, projectID string, key string) error
}

type secretsClient struct {
	*baseClient
}

// NewSecretsClient returns a specialized client for managing
// Secrets.
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
	opts meta.ListOptions,
) (SecretList, error) {
	secrets := SecretList{}
	return secrets, s.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			authHeaders: s.bearerTokenAuthHeaders(),
			queryParams: s.appendListQueryParams(nil, opts),
			successCode: http.StatusOK,
			respObj:     &secrets,
		},
	)
}

func (s *secretsClient) Set(
	ctx context.Context,
	projectID string,
	secret sdk.Secret,
) error {
	return s.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
			),
			authHeaders: s.bearerTokenAuthHeaders(),
			reqBodyObj:  secret,
			successCode: http.StatusOK,
		},
	)
}

func (s *secretsClient) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	return s.executeRequest(
		outboundRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
