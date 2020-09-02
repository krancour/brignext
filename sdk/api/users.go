package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// User represents a (human) BrigNext user.
type User struct {
	// ObjectMeta encapsulates User metadata.
	meta.ObjectMeta `json:"metadata"`
	// Name is the given name and surname of the User.
	Name string `json:"name,omitempty"`
	// Locked indicates when the User has been locked out of the system by
	// an administrator. If this field's value is nil, the User can be presumed
	// NOT to be locked.
	Locked *time.Time `json:"locked,omitempty"`
}

// MarshalJSON amends User instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "User",
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersSelector represents useful filter criteria when selecting multiple Users
// for API group operations like list. It currently has no fields, but exists
// for future expansion.
type UsersSelector struct{}

// UserList is an ordered and pageable list of Users.
type UserList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Users.
	Items []User `json:"items,omitempty"`
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
	Get(context.Context, string) (User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error
}

type usersClient struct {
	*baseClient
}

// NewUsersClient returns a specialized client for managing Users.
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
	_ context.Context,
	_ UsersSelector,
	opts meta.ListOptions,
) (UserList, error) {
	users := UserList{}
	return users, u.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/users",
			authHeaders: u.bearerTokenAuthHeaders(),
			queryParams: u.appendListQueryParams(nil, opts),
			successCode: http.StatusOK,
			respObj:     &users,
		},
	)
}

func (u *usersClient) Get(_ context.Context, id string) (User, error) {
	user := User{}
	return user, u.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/users/%s", id),
			authHeaders: u.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &user,
		},
	)
}

func (u *usersClient) Lock(_ context.Context, id string) error {
	return u.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/users/%s/lock", id),
			authHeaders: u.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (u *usersClient) Unlock(_ context.Context, id string) error {
	return u.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/users/%s/lock", id),
			authHeaders: u.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
