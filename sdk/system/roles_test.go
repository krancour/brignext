package system

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/v2/sdk/authx"
	"github.com/stretchr/testify/require"
)

const (
	testRole          = authx.RoleName("ceo")
	testUserID        = "tony@starkindustries.com"
	testPrincipalType = authx.PrincipalTypeUser
)

func TestNewRolesClient(t *testing.T) {
	client := NewRolesClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &rolesClient{}, client)
	requireBaseClient(t, client.(*rolesClient).BaseClient)
}

func TestRolesClientGrant(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/system/role-assignments", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				roleAssignment := authx.RoleAssignment{}
				err = json.Unmarshal(bodyBytes, &roleAssignment)
				require.NoError(t, err)
				require.Equal(t, testRole, roleAssignment.Role)
				require.Equal(t, testPrincipalType, roleAssignment.PrincipalType)
				require.Equal(t, testUserID, roleAssignment.PrincipalID)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRolesClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Grant(
		context.Background(),
		authx.RoleAssignment{
			Role:          testRole,
			PrincipalType: testPrincipalType,
			PrincipalID:   testUserID,
		},
	)
	require.NoError(t, err)
}

func TestRolesClientRevoke(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/system/role-assignments", r.URL.Path)
				require.Equal(t, testRole, authx.RoleName(r.URL.Query().Get("role")))
				require.Equal(
					t,
					testPrincipalType,
					authx.PrincipalType(r.URL.Query().Get("principalType")),
				)
				require.Equal(t, testUserID, r.URL.Query().Get("principalID"))
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewRolesClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Revoke(
		context.Background(),
		authx.RoleAssignment{
			Role:          testRole,
			PrincipalType: testPrincipalType,
			PrincipalID:   testUserID,
		},
	)
	require.NoError(t, err)
}
