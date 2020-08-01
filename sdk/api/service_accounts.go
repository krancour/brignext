package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// ServiceAccount represents a non-human BrigNext user, such as an Event
// gateway.
type ServiceAccount struct {
	// ObjectMeta encapsulates ServiceAccount metadata.
	meta.ObjectMeta `json:"metadata"`
	// Description is a natural language description of the ServiceAccount's
	// purpose.
	Description string `json:"description,omitempty"`
	// Locked indicates whether the ServiceAccount has been locked out of the
	// system by an administrator.
	Locked *time.Time `json:"locked,omitempty"`
}

// MarshalJSON amends ServiceAccount instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (s ServiceAccount) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccount
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccount",
			},
			Alias: (Alias)(s),
		},
	)
}

type ServiceAccountReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	Description              string     `json:"description,omitempty"`
	Locked                   *time.Time `json:"locked,omitempty"`
}

func (s ServiceAccountReference) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountReference",
			},
			Alias: (Alias)(s),
		},
	)
}

type ServiceAccountReferenceList struct {
	Items []ServiceAccountReference `json:"items,omitempty"`
}

func (s ServiceAccountReferenceList) MarshalJSON() ([]byte, error) {
	type Alias ServiceAccountReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ServiceAccountReferenceList",
			},
			Alias: (Alias)(s),
		},
	)
}

type ServiceAccountsClient interface {
	Create(context.Context, ServiceAccount) (Token, error)
	List(context.Context) (ServiceAccountReferenceList, error)
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
	return token, s.executeRequest(
		outboundRequest{
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
) (ServiceAccountReferenceList, error) {
	serviceAccountList := ServiceAccountReferenceList{}
	return serviceAccountList, s.executeRequest(
		outboundRequest{
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
	return serviceAccount, s.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/service-accounts/%s", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &serviceAccount,
		},
	)
}

func (s *serviceAccountsClient) Lock(_ context.Context, id string) error {
	return s.executeRequest(
		outboundRequest{
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
	return token, s.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &token,
		},
	)
}
