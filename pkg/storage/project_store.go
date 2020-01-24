package storage

import (
	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
)

type ProjectStore interface {
	CreateProject(project *brigade.Project) error
	GetProjects() ([]*brigade.Project, error)
	GetProject(id string) (*brigade.Project, error)
	UpdateProject(project *brigade.Project) error
	DeleteProject(id string) error

	CreateBuild(build *brigade.Build) error
	GetBuilds() ([]*brigade.Build, error)
	GetProjectBuilds(projectID string) ([]*brigade.Build, error)
	GetBuild(id string) (*brigade.Build, error)
	DeleteBuild(id string, options storage.DeleteBuildOptions) error

	UpdateWorker(worker *brigade.Worker) error

	CreateJob(buildID string, job *brigade.Job) error
	GetBuildJobs(buildID string) ([]*brigade.Job, error)
	UpdateJobStatus(jobID string, status string) error
	GetJob(id string) (*brigade.Job, error)
}
