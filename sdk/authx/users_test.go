package authx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/v2/sdk/meta"
	"github.com/stretchr/testify/require"
)

const testUserID = "tony@starkindustries.com"

func TestUserListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, UserList{}, "UserList")
}

func TestUserMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, User{}, "User")
}

func TestNewUsersClient(t *testing.T) {
	client := NewUsersClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &usersClient{}, client)
	requireBaseClient(t, client.(*usersClient).BaseClient)
}

func TestUsersClientList(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/users", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.List(
		context.Background(),
		UsersSelector{},
		meta.ListOptions{},
	)
	require.NoError(t, err)
}

func TestUsersClientGet(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s", testUserID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Get(context.Background(), testUserID)
	require.NoError(t, err)
}

func TestUsersClientLock(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s/lock", testUserID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Lock(context.Background(), testUserID)
	require.NoError(t, err)
}

func TestUsersClientUnlock(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/users/%s/lock", testUserID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewUsersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Unlock(context.Background(), testUserID)
	require.NoError(t, err)
}
