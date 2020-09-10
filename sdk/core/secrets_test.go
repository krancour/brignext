package core

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

const testSecretKey = "theooze"

func TestSecretListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, SecretList{}, "SecretList")
}

func TestSecretMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, Secret{}, "Secret")
}

func TestNewSecretsClient(t *testing.T) {
	client := NewSecretsClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &secretsClient{}, client)
	requireBaseClient(t, client.(*secretsClient).BaseClient)
}

func TestSecretsClientList(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s/secrets", testProjectID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.List(context.Background(), testProjectID, meta.ListOptions{})
	require.NoError(t, err)
}

func TestSecretsClientSet(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/projects/%s/secrets/%s",
						testProjectID,
						testSecretKey,
					),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				secret := Secret{}
				err = json.Unmarshal(bodyBytes, &secret)
				require.NoError(t, err)
				require.Equal(t, testSecretKey, secret.Key)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Set(
		context.Background(),
		testProjectID,
		Secret{
			Key: testSecretKey,
		},
	)
	require.NoError(t, err)
}

func TestSecretsClientUnset(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/projects/%s/secrets/%s",
						testProjectID,
						testSecretKey,
					),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewSecretsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Unset(context.Background(), testProjectID, testSecretKey)
	require.NoError(t, err)
}
