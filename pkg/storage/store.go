package storage

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
)

type Store interface {
	DoTx(context.Context, func(context.Context) error) error

	CreateUser(context.Context, brignext.User) error
	GetUsers(context.Context) ([]brignext.User, error)
	GetUser(context.Context, string) (brignext.User, bool, error)
	LockUser(context.Context, string) (bool, error)
	UnlockUser(context.Context, string) (bool, error)

	CreateSession(context.Context, brignext.Session) error
	GetSessionByHashedOAuth2State(
		context.Context,
		string,
	) (brignext.Session, bool, error)
	GetSessionByHashedToken(
		context.Context,
		string,
	) (brignext.Session, bool, error)
	AuthenticateSession(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) (bool, error)
	DeleteSession(context.Context, string) (bool, error)
	DeleteSessionsByUser(context.Context, string) (int64, error)

	CreateServiceAccount(context.Context, brignext.ServiceAccount) error
	GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error)
	GetServiceAccount(
		context.Context,
		string,
	) (brignext.ServiceAccount, bool, error)
	GetServiceAccountByHashedToken(
		context.Context,
		string,
	) (brignext.ServiceAccount, bool, error)
	LockServiceAccount(context.Context, string) (bool, error)
	UnlockServiceAccount(
		ctx context.Context,
		id string,
		newHashedToken string,
	) (bool, error)

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) ([]brignext.Project, error)
	GetProject(context.Context, string) (brignext.Project, bool, error)
	UpdateProject(context.Context, brignext.Project) (bool, error)
	DeleteProject(context.Context, string) (bool, error)

	CreateEvent(context.Context, brignext.Event) error
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, bool, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (int64, error)
}
