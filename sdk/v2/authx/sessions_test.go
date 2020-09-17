package authx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOIDCAuthDetailsMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(
		t,
		OIDCAuthDetails{},
		"OIDCAuthDetails",
	)
}

func TestNewSessionsClient(t *testing.T) {
	client := NewSessionsClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &sessionsClient{}, client)
	requireBaseClient(t, client.(*sessionsClient).BaseClient)
}

func TestSessionsClientCreateRootSession(t *testing.T) {
	const testRootPassword = "foobar"
	testSessionToken := Token{
		Value: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/sessions", r.URL.Path)
				require.Contains(t, r.Header.Get("Authorization"), "Basic")
				bodyBytes, err := json.Marshal(testSessionToken)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	token, err := client.CreateRootSession(context.Background(), testRootPassword)
	require.NoError(t, err)
	require.Equal(t, testSessionToken, token)
}

func TestSessionsClientCreateUserSession(t *testing.T) {
	testOIDCAuthDetails := OIDCAuthDetails{
		Token: "opensesame",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/sessions", r.URL.Path)
				require.Empty(t, r.Header.Get("Authorization"))
				bodyBytes, err := json.Marshal(testOIDCAuthDetails)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	OIDCAuthDetails, err := client.CreateUserSession(context.Background())
	require.NoError(t, err)
	require.Equal(t, testOIDCAuthDetails, OIDCAuthDetails)
}

func TestSessionsClientDelete(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/session", r.URL.Path)
				require.Contains(t, r.Header.Get("Authorization"), "Bearer")
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSessionsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Delete(context.Background())
	require.NoError(t, err)
}
