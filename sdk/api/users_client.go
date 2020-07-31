package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type UsersClient interface {
	List(context.Context) (UserReferenceList, error)
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

func (u *usersClient) List(
	context.Context,
) (UserReferenceList, error) {
	userList := UserReferenceList{}
	return userList, u.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &userList,
		},
	)
}

func (u *usersClient) Get(_ context.Context, id string) (User, error) {
	user := User{}
	return user, u.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersClient) Unlock(_ context.Context, id string) error {
	return u.ExecuteRequest(
		OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
