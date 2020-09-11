package authx

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
	"github.com/brigadecore/brigade/v2/sdk/meta"
)

type UserSessionAuthDetails struct {
	OAuth2State string `json:"oauth2State,omitempty"`
	AuthURL     string `json:"authURL,omitempty"`
	Token       string `json:"token,omitempty"`
}

// MarshalJSON amends UserSessionAuthDetails instances with type metadata so
// that clients do not need to be concerned with the tedium of doing so.
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

// SessionsClient is the specialized client for managing Brigade API Sessions.
type SessionsClient interface {
	CreateRootSession(
		ctx context.Context,
		password string,
	) (Token, error)
	CreateUserSession(context.Context) (UserSessionAuthDetails, error)
	Delete(context.Context) error
}

type sessionsClient struct {
	*restmachinery.BaseClient
}

// NewSessionsClient returns a specialized client for managing Brigade API
// Sessions.
func NewSessionsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SessionsClient {
	return &sessionsClient{
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

func (s *sessionsClient) CreateRootSession(
	ctx context.Context,
	password string,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			AuthHeaders: s.BasicAuthHeaders("root", password),
			QueryParams: map[string]string{
				"root": "true",
			},
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (s *sessionsClient) CreateUserSession(
	ctx context.Context,
) (UserSessionAuthDetails, error) {
	userSessionAuthDetails := UserSessionAuthDetails{}
	return userSessionAuthDetails, s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			SuccessCode: http.StatusCreated,
			RespObj:     &userSessionAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(ctx context.Context) error {
	return s.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
