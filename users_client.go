package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type UsersClient interface {
	List(context.Context) (UserList, error)
	Get(context.Context, string) (User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}

type usersClient struct {
	*baseClient
}

func NewUsersClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) UsersClient {
	return &usersClient{
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

func (c *usersClient) List(context.Context) (UserList, error) {
	userList := UserList{}
	err := c.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/users",
			authHeaders: c.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &userList,
		},
	)
	return userList, err
}

func (c *usersClient) Get(_ context.Context, id string) (User, error) {
	user := User{}
	err := c.executeAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/users/%s", id),
			authHeaders: c.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &user,
		},
	)
	return user, err
}

func (c *usersClient) Lock(_ context.Context, id string) error {
	return c.executeAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/users/%s/lock", id),
			authHeaders: c.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (c *usersClient) Unlock(_ context.Context, id string) error {
	return c.executeAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/users/%s/lock", id),
			authHeaders: c.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
