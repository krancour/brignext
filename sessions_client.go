package brignext

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type SessionsClient interface {
	CreateRootSession(ctx context.Context, password string) (Token, error)
	CreateUserSession(context.Context) (UserSessionAuthDetails, error)
	Delete(context.Context) error
}

type sessionsClient struct {
	*api.BaseClient
}

func NewSessionsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) SessionsClient {
	return &sessionsClient{
		BaseClient: &api.BaseClient{
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
	_ context.Context,
	password string,
) (Token, error) {
	token := Token{}
	return token, s.ExecuteRequest(
		api.Request{
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
	context.Context,
) (UserSessionAuthDetails, error) {
	userSessionAuthDetails := UserSessionAuthDetails{}
	return userSessionAuthDetails, s.ExecuteRequest(
		api.Request{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			SuccessCode: http.StatusCreated,
			RespObj:     &userSessionAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(context.Context) error {
	return s.ExecuteRequest(
		api.Request{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
