package sessions

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type Client interface {
	CreateRootSession(ctx context.Context, password string) (brignext.Token, error)
	CreateUserSession(context.Context) (brignext.UserSessionAuthDetails, error)
	Delete(context.Context) error
}

type client struct {
	*api.BaseClient
}

func NewClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) Client {
	return &client{
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

func (c *client) CreateRootSession(
	_ context.Context,
	password string,
) (brignext.Token, error) {
	token := brignext.Token{}
	return token, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			AuthHeaders: c.BasicAuthHeaders("root", password),
			QueryParams: map[string]string{
				"root": "true",
			},
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (c *client) CreateUserSession(
	context.Context,
) (brignext.UserSessionAuthDetails, error) {
	userSessionAuthDetails := brignext.UserSessionAuthDetails{}
	return userSessionAuthDetails, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/sessions",
			SuccessCode: http.StatusCreated,
			RespObj:     &userSessionAuthDetails,
		},
	)
}

func (c *client) Delete(context.Context) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/session",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
