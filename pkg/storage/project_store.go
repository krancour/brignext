package storage

import (
	"github.com/krancour/brignext/pkg/brignext"
)

type ProjectStore interface {
	CreateProject(project brignext.Project) (string, error)
	GetProjects() ([]brignext.Project, error)
	GetProject(id string) (brignext.Project, bool, error)
	UpdateProject(project brignext.Project) error
	DeleteProject(id string) error

	CreateEvent(event brignext.Event) (string, error)
	GetEvents() ([]brignext.Event, error)
	GetEventsByProjectID(projectID string) ([]brignext.Event, error)
	GetEvent(id string) (brignext.Event, bool, error)
	DeleteEventsByProjectID(projectID string, options DeleteEventOptions) error
	DeleteEvent(id string, options DeleteEventOptions) error

	UpdateWorker(eventID string, worker brignext.Worker) error

	CreateJob(job brignext.Job) (string, error)
	GetJobsByEventID(eventID string) ([]brignext.Job, error)
	GetJob(id string) (brignext.Job, bool, error)
	UpdateJobStatus(jobID string, status string) error
	DeleteJobsByEventID(eventID string) error
}

type DeleteEventOptions struct {
	DeleteEventsWithRunningWorkers bool
}
