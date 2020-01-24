package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: These tests are super quick and dirty and need to be run in a certain
// order. These helped me during development, but they're crap and I am not
// proud of them. These need a lot of love when time permits. They are NOT
// suitable as is... and they may actually be broken.

func TestCreateProject(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.CreateProject(
		&brigade.Project{
			ID:   "foo",
			Name: "bar",
		},
	)
	require.NoError(t, err)
}

func TestGetProjects(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	projects, err := store.GetProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "foo", projects[0].ID)
	require.Equal(t, "bar", projects[0].Name)
}

func TestGetProject(t *testing.T) {
	const id = "foo"
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	project, err := store.GetProject(id)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, id, project.ID)
	require.Equal(t, "bar", project.Name)
}

func TestUpdateProject(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.UpdateProject(
		&brigade.Project{
			ID:   "foo",
			Name: "bat",
		},
	)
	require.NoError(t, err)
}

func TestDeleteProject(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.DeleteProject("foo")
	require.NoError(t, err)
}

func TestCreateBuild(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.CreateBuild(
		&brigade.Build{
			ID:        "foo-build-1",
			ProjectID: "foo",
			Worker: &brigade.Worker{
				ID: "foo-worker",
			},
		},
	)
	require.NoError(t, err)
}

func TestUpdateWorker(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.UpdateWorker(
		&brigade.Worker{
			BuildID: "foo-build-1",
			Status:  "Succeeded",
		},
	)
	require.NoError(t, err)
}

func TestGetBuilds(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	builds, err := store.GetBuilds()
	require.NoError(t, err)
	require.Len(t, builds, 1)
	require.Equal(t, "foo-build-1", builds[0].ID)
	require.Equal(t, "foo", builds[0].ProjectID)
}

func TestGetProjectBuilds(t *testing.T) {
	const projectID = "foo"
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	builds, err := store.GetProjectBuilds(projectID)
	require.NoError(t, err)
	require.Len(t, builds, 1)
	require.Equal(t, "foo-build-1", builds[0].ID)
	require.Equal(t, projectID, builds[0].ProjectID)
}

func TestGetBuild(t *testing.T) {
	const id = "foo-build-1"
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	build, err := store.GetBuild(id)
	require.NoError(t, err)
	require.NotNil(t, build)
	require.Equal(t, id, build.ID)
	require.Equal(t, "foo", build.ProjectID)
}

func TestDeleteBuild(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.DeleteBuild(
		"foo-build-1",
		storage.DeleteBuildOptions{
			SkipRunningBuilds: true,
		},
	)
	require.NoError(t, err)
}

func TestCreateJob(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db).(*projectStore)
	err = store.CreateJob(
		"foo-build-1",
		&brigade.Job{
			ID: "foo-job-1",
		},
	)
	require.NoError(t, err)
}

func TestUpdateJobStatus(t *testing.T) {
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	err = store.UpdateJobStatus("foo-job-1", "wooooooo")
	require.NoError(t, err)
}

func TestGetBuildJobs(t *testing.T) {
	const buildID = "foo-build-1"
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	jobs, err := store.GetBuildJobs(buildID)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	require.Equal(t, "foo-job-1", jobs[0].ID)
}

func TestGetJob(t *testing.T) {
	const id = "foo-job-1"
	db, err := getTestDatabase()
	require.NoError(t, err)
	store := NewProjectStore(db)
	job, err := store.GetJob(id)
	require.NoError(t, err)
	require.NotNil(t, job)
	require.Equal(t, id, job.ID)
}

func getTestDatabase() (*mongo.Database, error) {
	connectCtx, connectCancel :=
		context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()
	client, err := mongo.Connect(
		connectCtx,
		options.Client().ApplyURI(
			fmt.Sprintf(
				"mongodb://%s:%s@%s:%d/%s",
				"fluentd",
				"foobar",
				"52.147.214.165",
				27017,
				"logs",
			),
		),
	)
	if err != nil {
		return nil, err
	}
	return client.Database("logs"), nil
}
