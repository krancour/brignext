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
	DeleteEvents(criteria DeleteEventsCriteria) error

	// CreateWorker(worker brignext.Worker) (string, error)
	// GetWorkers(criteria GetWorkersCriteria) ([]brignext.Worker, error)
	// GetWorker(id string) (brignext.Worker, error)
	// CancelWorkers(criteria CancelWorkersCriteria) error
	// DeleteWorkers(criteria DeleteWorkersCriteria) error
}

type GetEventsCriteria struct {
	ProjectID string
}

type DeleteEventsCriteria struct {
	ProjectID                      string
	EventID                        string
	DeleteEventsWithPendingWorkers bool
	DeleteEventsWithRunningWorkers bool
}

// type GetWorkersCriteria struct {
// 	ProjectID string
// 	EventID  string
// }

// type CancelWorkersCriteria struct {
// 	ProjectID           string
// 	EventID            string
// 	WorkerID           string
// 	StopRunningWorkers bool
// }

// type DeleteWorkersCriteria struct {
// 	ProjectID             string
// 	EventID              string
// 	WorkerID             string
// 	DeletePendingWorkers bool
// 	DeleteRunningWorkers bool
// }
