package core

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestNewProjectRolesClient(t *testing.T) {
	client := NewProjectRolesClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &projectRolesClient{}, client)
	requireBaseClient(t, client.(*projectRolesClient).BaseClient)
}

func TestProjectRolesClientGrant(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
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
	client := NewProjectRolesClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Grant(
		context.Background(),
		testProjectID,
		authx.RoleAssignment{
			Role:          testRole,
			PrincipalType: testPrincipalType,
			PrincipalID:   testUserID,
		},
	)
	require.NoError(t, err)
}

func TestProjectRolesClientRevoke(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/role-assignments", testProjectID),
					r.URL.Path,
				)
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
	client := NewProjectRolesClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Revoke(
		context.Background(),
		testProjectID,
		authx.RoleAssignment{
			Role:          testRole,
			PrincipalType: testPrincipalType,
			PrincipalID:   testUserID,
		},
	)
	require.NoError(t, err)
}
