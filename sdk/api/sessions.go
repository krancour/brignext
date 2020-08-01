package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type UserSessionAuthDetails struct {
	OAuth2State string `json:"oauth2State,omitempty"`
	AuthURL     string `json:"authURL,omitempty"`
	Token       string `json:"token,omitempty"`
}

func (u UserSessionAuthDetails) MarshalJSON() ([]byte, error) {
	type Alias UserSessionAuthDetails
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserSessionAuthDetails",
			},
			Alias: (Alias)(u),
		},
	)
}

// SessionsClient is the specialized interface for managing BrigNext API
// Sessions.
type SessionsClient interface {
	CreateRootSession(
		ctx context.Context,
		password string,
	) (Token, error)
	CreateUserSession(context.Context) (UserSessionAuthDetails, error)
	Delete(context.Context) error
}

type sessionsClient struct {
	*baseClient
}

// NewSessionsClient returns a specialized client for managing BrigNext API
// Sessions.
func NewSessionsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SessionsClient {
	return &sessionsClient{
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

func (s *sessionsClient) CreateRootSession(
	_ context.Context,
	password string,
) (Token, error) {
	token := Token{}
	return token, s.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/sessions",
			authHeaders: s.basicAuthHeaders("root", password),
			queryParams: map[string]string{
				"root": "true",
			},
			successCode: http.StatusCreated,
			respObj:     &token,
		},
	)
}

func (s *sessionsClient) CreateUserSession(
	context.Context,
) (UserSessionAuthDetails, error) {
	userSessionAuthDetails := UserSessionAuthDetails{}
	return userSessionAuthDetails, s.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/sessions",
			successCode: http.StatusCreated,
			respObj:     &userSessionAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(context.Context) error {
	return s.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        "v2/session",
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
