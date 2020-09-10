package authx

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/v2/sdk/meta"
	"github.com/stretchr/testify/require"
)

const testServiceAccountID = "jarvis"
const testServiceAccountToken = "opensesame"

func TestServiceAccountListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, ServiceAccountList{}, "ServiceAccountList")
}

func TestServiceMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, ServiceAccount{}, "ServiceAccount")
}

func TestNewServiceAccountsClient(t *testing.T) {
	client := NewServiceAccountsClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &serviceAccountsClient{}, client)
	requireBaseClient(t, client.(*serviceAccountsClient).BaseClient)
}

func TestServiceAccountsClientCreate(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/service-accounts", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				serviceAccount := ServiceAccount{}
				err = json.Unmarshal(bodyBytes, &serviceAccount)
				require.NoError(t, err)
				require.Equal(t, testServiceAccountID, serviceAccount.ID)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, `{"value":%q}`, testServiceAccountToken)
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	token, err := client.Create(
		context.Background(),
		ServiceAccount{
			ObjectMeta: meta.ObjectMeta{
				ID: testServiceAccountID,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, testServiceAccountToken, token.Value)
}

func TestServiceAccountsClientList(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/service-accounts", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.List(
		context.Background(),
		ServiceAccountsSelector{},
		meta.ListOptions{},
	)
	require.NoError(t, err)
}

func TestServiceAccountsClientGet(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s", testServiceAccountID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Get(context.Background(), testServiceAccountID)
	require.NoError(t, err)
}

func TestServiceAccountsClientLock(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s/lock", testServiceAccountID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Lock(context.Background(), testServiceAccountID)
	require.NoError(t, err)
}

func TestServiceAccountsClientUnlock(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/service-accounts/%s/lock", testServiceAccountID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"value":%q}`, testServiceAccountToken)
			},
		),
	)
	defer server.Close()
	client := NewServiceAccountsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	token, err := client.Unlock(context.Background(), testServiceAccountID)
	require.NoError(t, err)
	require.Equal(t, testServiceAccountToken, token.Value)
}
