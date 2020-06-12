package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type UsersClient interface {
	List(context.Context) (UserList, error)
	Get(context.Context, string) (User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}

type usersClient struct {
	*api.BaseClient
}

func NewUsersClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) UsersClient {
	return &usersClient{
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

func (c *usersClient) List(context.Context) (UserList, error) {
	userList := UserList{}
	return userList, c.ExecuteRequest(
		api.Request{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &userList,
		},
	)
}

func (c *usersClient) Get(_ context.Context, id string) (User, error) {
	user := User{}
	return user, c.ExecuteRequest(
		api.Request{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/users/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &user,
		},
	)
}

func (c *usersClient) Lock(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.Request{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *usersClient) Unlock(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.Request{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
