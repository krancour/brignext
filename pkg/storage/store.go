package storage

import (
	"github.com/krancour/brignext/pkg/brignext"
)

type Store interface {
	CreateUser(user brignext.User) error
	GetUsers() ([]brignext.User, error)
	GetUser(id string) (brignext.User, bool, error)
	LockUser(id string) error
	UnlockUser(id string) error

	CreateSession(session brignext.Session) (string, string, string, error)
	GetSession(criteria GetSessionCriteria) (brignext.Session, bool, error)
	AuthenticateSession(sessionID, userID string) error
	DeleteSessions(criteria DeleteSessionsCriteria) error

	CreateServiceAccount(serviceAccount brignext.ServiceAccount) (string, error)
	GetServiceAccounts() ([]brignext.ServiceAccount, error)
	GetServiceAccount(
		criteria GetServiceAccountCriteria,
	) (brignext.ServiceAccount, bool, error)
	LockServiceAccount(id string) error
	UnlockServiceAccount(id string) (string, error)

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

type GetSessionCriteria struct {
	OAuth2State string
	Token       string
}

type DeleteSessionsCriteria struct {
	SessionID string
	UserID    string
}

type GetServiceAccountCriteria struct {
	ServiceAccountID string
	Token            string
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
