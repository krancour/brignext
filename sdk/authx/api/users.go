package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/authx"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// UsersSelector represents useful filter criteria when selecting multiple Users
// for API group operations like list. It currently has no fields, but exists
// for future expansion.
type UsersSelector struct{}

// UserList is an ordered and pageable list of Users.
type UserList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Users.
	Items []authx.User `json:"items,omitempty"`
}

// MarshalJSON amends UserList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (u UserList) MarshalJSON() ([]byte, error) {
	type Alias UserList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserList",
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersClient is the specialized client for managing Users with the BrigNext
// API.
type UsersClient interface {
	// List returns a UserList.
	List(context.Context, UsersSelector, meta.ListOptions) (UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (authx.User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error
}

type usersClient struct {
	*apimachinery.BaseClient
}

// NewUsersClient returns a specialized client for managing Users.
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
	_ context.Context,
	_ UsersSelector,
	opts meta.ListOptions,
) (UserList, error) {
	users := UserList{}
	return users, u.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: u.BearerTokenAuthHeaders(),
			QueryParams: u.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &users,
		},
	)
}

func (u *usersClient) Get(_ context.Context, id string) (authx.User, error) {
	user := authx.User{}
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
