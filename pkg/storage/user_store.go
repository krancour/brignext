package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
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
