package brignext

import (
	"context"
	"crypto/tls"
	"net/http"
)

type SessionsClient interface {
	CreateRootSession(ctx context.Context, password string) (Token, error)
	CreateUserSession(context.Context) (UserSessionAuthDetails, error)
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
) (Token, error) {
	token := Token{}
	return token, s.executeAPIRequest(
		apiRequest{
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
	return userSessionAuthDetails, s.executeAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/sessions",
			successCode: http.StatusCreated,
			respObj:     &userSessionAuthDetails,
		},
	)
}

func (s *sessionsClient) Delete(context.Context) error {
	return s.executeAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        "v2/session",
			authHeaders: s.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
