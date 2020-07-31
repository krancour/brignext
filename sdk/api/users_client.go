package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
)

type UsersClient interface {
	List(context.Context) (brignext.UserReferenceList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}

type usersClient struct {
	*apimachinery.BaseClient
}

func NewUsersClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) UsersClient {
	return &usersClient{
		BaseClient: &apimachinery.BaseClient{
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

func (u *usersClient) List(
	context.Context,
) (brignext.UserReferenceList, error) {
	userList := brignext.UserReferenceList{}
	return userList, u.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &userList,
		},
	)
}

func (u *usersClient) Get(_ context.Context, id string) (brignext.User, error) {
	user := brignext.User{}
	return user, u.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/users/%s", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &user,
		},
	)
}

func (u *usersClient) Lock(_ context.Context, id string) error {
	return u.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersClient) Unlock(_ context.Context, id string) error {
	return u.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
