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
	// Locked indicates whether the User has been locked out of the system by
	// an administrator.
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

type UserReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	Name                     string     `json:"name,omitempty"`
	Locked                   *time.Time `json:"locked,omitempty"`
}

func (u UserReference) MarshalJSON() ([]byte, error) {
	type Alias UserReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserReference",
			},
			Alias: (Alias)(u),
		},
	)
}

type UserReferenceList struct {
	Items []UserReference `json:"items,omitempty"`
}

func (u UserReferenceList) MarshalJSON() ([]byte, error) {
	type Alias UserReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserReferenceList",
			},
			Alias: (Alias)(u),
		},
	)
}

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
	return userList, u.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/users",
			authHeaders: u.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &userList,
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
