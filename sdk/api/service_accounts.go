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
	// Locked indicates when the ServiceAccount has been locked out of the system
	// by an administrator. If this field's value is nil, the User can be presumed
	// NOT to be locked.
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

// ServiceAccountReference is an abridged representation of a ServiceAccount
// useful to API operations that construct and return potentially large
// collections of service accounts.
type ServiceAccountReference struct {
	// ObjectReferenceMeta contains abridged ServiceAccount metadata.
	meta.ObjectReferenceMeta `json:"metadata"`
	// Description is a natural language description of the ServiceAccount's
	// purpose.
	Description string `json:"description,omitempty"`
	// Locked indicates when the ServiceAccount has been locked out of the system
	// by an administrator. If this field's value is nil, the User can be presumed
	// NOT to be locked.
	Locked *time.Time `json:"locked,omitempty"`
}

// MarshalJSON amends ServiceAccountReference instances with type metadata so
// that clients do not need to be concerned with the tedium of doing so.
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

// ServiceAccountReferenceList is an ordered list of ServiceAccountReferences.
type ServiceAccountReferenceList struct {
	// Items is a slice of ServiceAccountReferences.
	//
	// TODO: When pagination is implemented, list metadata will need to be added
	Items []ServiceAccountReference `json:"items,omitempty"`
}

// MarshalJSON amends ServiceAccountReferenceList instances with type metadata
// so that clients do not need to be concerned with the tedium of doing so.
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

// ServiceAccountsClient is the specialized client for managing ServiceAccounts
// with the BrigNext API.
type ServiceAccountsClient interface {
	// Create creates a new ServiceAccount.
	Create(context.Context, ServiceAccount) (Token, error)
	// List returns a ServiceAccountReferenceList.
	//
	// TODO: This should take some list options because we may want them in the
	// future and they would be hard to add later.
	List(context.Context) (ServiceAccountReferenceList, error)
	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (ServiceAccount, error)
	// Lock removes access to the API for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single ServiceAccount specified by
	// its identifier. It returns a new Token.
	Unlock(context.Context, string) (Token, error)
}

type serviceAccountsClient struct {
	*baseClient
}

// NewServiceAccountsClient returns a specialized client for managing
// ServiceAccounts.
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
