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

const testProjectID = "bluebook"

func TestProjectListMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, ProjectList{}, "ProjectList")
}

func TestProjectMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, Project{}, "Project")
}

func TestNewProjectsClient(t *testing.T) {
	client := NewProjectsClient(
		testAPIAddress,
		testAPIToken,
		testClientAllowInsecure,
	)
	require.IsType(t, &projectsClient{}, client)
	requireBaseClient(t, client.(*projectsClient).BaseClient)
	require.NotNil(t, client.(*projectsClient).rolesClient)
	require.Equal(t, client.(*projectsClient).rolesClient, client.Roles())
	require.NotNil(t, client.(*projectsClient).secretsClient)
	require.Equal(t, client.(*projectsClient).secretsClient, client.Secrets())
}

func TestProjectsClientCreate(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProjectID, project.ID)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Create(
		context.Background(),
		Project{
			ObjectMeta: meta.ObjectMeta{
				ID: testProjectID,
			},
		},
	)
	require.NoError(t, err)
}

func TestProjectsClientCreateFromBytes(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProjectID, project.ID)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.CreateFromBytes(
		context.Background(),
		[]byte(fmt.Sprintf(`{"metadata":{"id":%q}}`, testProjectID)),
	)
	require.NoError(t, err)
}

func TestProjectsClientList(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/projects", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.List(
		context.Background(),
		ProjectsSelector{},
		meta.ListOptions{},
	)
	require.NoError(t, err)
}

func TestProjectsClientGet(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProjectID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Get(context.Background(), testProjectID)
	require.NoError(t, err)
}

func TestProjectsClientUpdate(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProjectID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProjectID, project.ID)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.Update(
		context.Background(),
		Project{
			ObjectMeta: meta.ObjectMeta{
				ID: testProjectID,
			},
		},
	)
	require.NoError(t, err)
}

func TestProjectsClientUpdateFromBytes(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProjectID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				project := Project{}
				err = json.Unmarshal(bodyBytes, &project)
				require.NoError(t, err)
				require.Equal(t, testProjectID, project.ID)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	_, err := client.UpdateFromBytes(
		context.Background(),
		testProjectID,
		[]byte(fmt.Sprintf(`{"metadata":{"id":%q}}`, testProjectID)),
	)
	require.NoError(t, err)
}

func TestProjectsClientDelete(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/projects/%s", testProjectID),
					r.URL.Path,
				)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewProjectsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Delete(context.Background(), testProjectID)
	require.NoError(t, err)
}
