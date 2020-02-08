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

	CreateEvent(event brignext.Event) error
	GetEvents() ([]brignext.Event, error)
	GetEventsByProjectName(projectName string) ([]brignext.Event, error)
	GetEvent(id string) (brignext.Event, bool, error)
	DeleteEvent(id string, options DeleteEventOptions) error

	UpdateWorker(eventID string, worker brignext.Worker) error

	CreateJob(job brignext.Job) error
	GetJobsByEventID(eventID string) ([]brignext.Job, error)
	GetJob(id string) (brignext.Job, bool, error)
	UpdateJobStatus(jobID string, status string) error
	DeleteJobsByEventID(eventID string) error
}

type DeleteEventOptions struct {
	DeleteEventsWithRunningWorkers bool
}
