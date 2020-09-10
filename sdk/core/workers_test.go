package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkerStatusMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, WorkerStatus{}, "WorkerStatus")
}

func TestNewWorkersClient(t *testing.T) {
	client := NewWorkersClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &workersClient{}, client)
	requireBaseClient(t, client.(*workersClient).BaseClient)
	require.NotNil(t, client.(*workersClient).jobsClient)
	require.Equal(t, client.(*workersClient).jobsClient, client.Jobs())
}

func TestWorkersClientStart(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/start", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Start(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestWorkersClientGetStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.GetStatus(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestWorkersClientWatchStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				require.Equal(
					t,
					"true",
					r.URL.Query().Get("watch"),
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, _, err := client.WatchStatus(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestWorkersClientUpdateStatus(t *testing.T) {
	const testPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				workerStatus := WorkerStatus{}
				err = json.Unmarshal(bodyBytes, &workerStatus)
				require.NoError(t, err)
				require.Equal(t, testPhase, workerStatus.Phase)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.UpdateStatus(
		context.Background(),
		testEventID,
		WorkerStatus{
			Phase: testPhase,
		},
	)
	require.NoError(t, err)
}
