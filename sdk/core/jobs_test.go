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

const testJobName = "Italian"

func TestJobSpecMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, JobSpec{}, "JobSpec")
}

func TestJobStatusMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, JobStatus{}, "JobStatus")
}

func TestJobsClientStart(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s/start",
						testEventID,
						testJobName,
					),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Start(context.Background(), testEventID, testJobName)
	require.NoError(t, err)
}

func TestJobsClientGetStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/jobs/%s/status", testEventID, testJobName),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.GetStatus(context.Background(), testEventID, testJobName)
	require.NoError(t, err)
}

func TestJobsClientWatchStatus(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/jobs/%s/status", testEventID, testJobName),
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
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, _, err := client.WatchStatus(
		context.Background(),
		testEventID,
		testJobName,
	)
	require.NoError(t, err)
}

func TestJobClientUpdateStatus(t *testing.T) {
	const testPhase = JobPhaseRunning
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/jobs/%s/status", testEventID, testJobName),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				jobStatus := JobStatus{}
				err = json.Unmarshal(bodyBytes, &jobStatus)
				require.NoError(t, err)
				require.Equal(t, testPhase, jobStatus.Phase)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.UpdateStatus(
		context.Background(),
		testEventID,
		testJobName,
		JobStatus{
			Phase: testPhase,
		},
	)
	require.NoError(t, err)
}
