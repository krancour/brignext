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
	GetEvents(criteria GetEventsCriteria) ([]brignext.Event, error)
	GetEvent(id string) (brignext.Event, bool, error)
	// TODO:
	// CancelEvents(criteria CancelEventsCriteria) error
	DeleteEvents(criteria DeleteEventsCriteria) error

	// TODO:
	// CancelWorker(criteria CancelWorkerCriteria) error
}

type GetEventsCriteria struct {
	ProjectID string
}

// type CancelEventsCriteria struct {
// 	ProjectID             string
// 	EventID               string
// 	AbortProcessingEvents bool
// }

type DeleteEventsCriteria struct {
	ProjectID              string
	EventID                string
	DeleteAcceptedEvents   bool
	DeleteProcessingEvents bool
}

// type CancelWorkerCriteria struct {
// 	EventID             string
// 	WorkerName          string
// 	AbortRunningWorkers bool
// }
