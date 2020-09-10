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

const (
	testEventID      = "1234567890"
	testEventPayload = "a Tesla roadster"
)

func TestEventListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, EventList{}, "EventList")
}

func TestListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, Event{}, "Event")
}

func TestNewEventsClient(t *testing.T) {
	client := NewEventsClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &eventsClient{}, client)
	requireBaseClient(t, client.(*eventsClient).BaseClient)
	require.NotNil(t, client.(*eventsClient).workersClient)
	require.Equal(t, client.(*eventsClient).workersClient, client.Workers())
	require.NotNil(t, client.(*eventsClient).logsClient)
	require.Equal(t, client.(*eventsClient).logsClient, client.Logs())
}

func TestEventsClientCreate(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/events", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				event := Event{}
				err = json.Unmarshal(bodyBytes, &event)
				require.NoError(t, err)
				require.Equal(t, testEventPayload, event.Payload)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Create(
		context.Background(),
		Event{
			Payload: testEventPayload,
		},
	)
	require.NoError(t, err)
}

func TestEventsClientList(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/events", r.URL.Path)
				require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
				require.Equal(
					t,
					testWorkerPhase,
					WorkerPhase(r.URL.Query().Get("workerPhases")),
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.List(
		context.Background(),
		EventsSelector{
			ProjectID:    testProjectID,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
		meta.ListOptions{},
	)
	require.NoError(t, err)
}

func TestEventsClientGet(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Get(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestEventsClientCancel(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/cancellation", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Cancel(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestEventsClientCancelMany(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/events/cancellations", r.URL.Path)
				require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
				require.Equal(
					t,
					testWorkerPhase,
					WorkerPhase(r.URL.Query().Get("workerPhases")),
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.CancelMany(
		context.Background(),
		EventsSelector{
			ProjectID:    testProjectID,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
	)
	require.NoError(t, err)
}

func TestEventsClientDelete(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Delete(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestEventsClientDeleteMany(t *testing.T) {
	const testWorkerPhase = WorkerPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/events", r.URL.Path)
				require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
				require.Equal(
					t,
					testWorkerPhase,
					WorkerPhase(r.URL.Query().Get("workerPhases")),
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.DeleteMany(
		context.Background(),
		EventsSelector{
			ProjectID:    testProjectID,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
	)
	require.NoError(t, err)
}
