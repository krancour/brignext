package storage

import (
	"github.com/krancour/brignext/pkg/brignext"
)

type ProjectStore interface {
	CreateProject(project brignext.Project) error
	GetProjects() ([]brignext.Project, error)
	GetProject(name string) (brignext.Project, bool, error)
	UpdateProject(project brignext.Project) error
	DeleteProject(name string) error

	CreateBuild(build brignext.Build) error
	GetBuilds() ([]brignext.Build, error)
	GetBuildsByProjectName(projectName string) ([]brignext.Build, error)
	GetBuild(id string) (brignext.Build, bool, error)
	DeleteBuild(id string, options DeleteBuildOptions) error

	UpdateWorker(buildID string, worker brignext.Worker) error

	CreateJob(job brignext.Job) error
	GetJobsByBuildID(buildID string) ([]brignext.Job, error)
	GetJob(id string) (brignext.Job, bool, error)
	UpdateJobStatus(jobID string, status string) error
	DeleteJobsByBuildID(buildID string) error
}

type DeleteBuildOptions struct {
	DeleteRunningBuilds bool
}
