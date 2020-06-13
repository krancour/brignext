package users

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/internal/api"
	brignext "github.com/krancour/brignext/v2/sdk"
)

type Client interface {
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
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

func (c *client) List(context.Context) (brignext.UserList, error) {
	userList := brignext.UserList{}
	return userList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &userList,
		},
	)
}

func (c *client) Get(_ context.Context, id string) (brignext.User, error) {
	user := brignext.User{}
	return user, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/users/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &user,
		},
	)
}

func (c *client) Lock(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) Unlock(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
