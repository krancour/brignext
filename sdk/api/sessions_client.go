package api

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
)

type SessionsClient interface {
	CreateRootSession(
		ctx context.Context,
		password string,
	) (sdk.Token, error)
	CreateUserSession(context.Context) (sdk.UserSessionAuthDetails, error)
	Delete(context.Context) error
}

type sessionsClient struct {
	*baseClient
}

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
) (sdk.Token, error) {
	token := sdk.Token{}
	return token, s.ExecuteRequest(
		OutboundRequest{
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
) (sdk.UserSessionAuthDetails, error) {
	userSessionAuthDetails := sdk.UserSessionAuthDetails{}
	return userSessionAuthDetails, s.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			SuccessCode: http.StatusCreated,
			RespObj:     &userSessionAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(context.Context) error {
	return s.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			AuthHeaders: s.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
