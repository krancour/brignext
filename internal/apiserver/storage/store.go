package storage

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
)

type Store interface {
	DoTx(context.Context, func(context.Context) error) error

	CreateUser(context.Context, brignext.User) error
	GetUsers(context.Context) ([]brignext.User, error)
	GetUser(context.Context, string) (brignext.User, error)
	LockUser(context.Context, string) error
	UnlockUser(context.Context, string) error

	CreateSession(context.Context, auth.Session) error
	GetSessionByHashedOAuth2State(
		context.Context,
		string,
	) (auth.Session, error)
	GetSessionByHashedToken(context.Context, string) (auth.Session, error)
	AuthenticateSession(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	DeleteSession(context.Context, string) error
	DeleteSessionsByUser(context.Context, string) (int64, error)

	CreateServiceAccount(context.Context, brignext.ServiceAccount) error
	GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error)
	GetServiceAccount(context.Context, string) (brignext.ServiceAccount, error)
	GetServiceAccountByHashedToken(
		context.Context,
		string,
	) (brignext.ServiceAccount, error)
	LockServiceAccount(context.Context, string) error
	UnlockServiceAccount(
		ctx context.Context,
		id string,
		newHashedToken string,
	) error

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) ([]brignext.Project, error)
	GetSubscribedProjects(
		ctx context.Context,
		event brignext.Event,
	) ([]brignext.Project, error)
	GetProject(context.Context, string) (brignext.Project, error)
	UpdateProject(context.Context, brignext.Project) error
	DeleteProject(context.Context, string) error

	CreateEvent(context.Context, brignext.Event) error
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, error)
	CancelEvent(
		ctx context.Context,
		id string,
		cancelRunning bool,
	) (bool, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteRunning bool,
	) (bool, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error

	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error
}
