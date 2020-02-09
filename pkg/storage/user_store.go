package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
	CreateUser(user brignext.User) error
	GetUsers() ([]brignext.User, error)
	GetUser(id string) (brignext.User, bool, error)
	LockUser(id string) error
	UnlockUser(id string) error

	CreateSession(session brignext.Session) (string, string, string, error)
	GetSessionByOAuth2State(oauth2State string) (brignext.Session, bool, error)
	GetSessionByToken(token string) (brignext.Session, bool, error)
	AuthenticateSession(sessionID, userID string) error
	DeleteSession(id string) error
	DeleteSessionsByUserID(userID string) error

	CreateServiceAccount(serviceAccount brignext.ServiceAccount) (string, error)
	GetServiceAccounts() ([]brignext.ServiceAccount, error)
	GetServiceAccount(id string) (brignext.ServiceAccount, bool, error)
	GetServiceAccountByToken(token string) (brignext.ServiceAccount, bool, error)
	LockServiceAccount(id string) error
	UnlockServiceAccount(id string) (string, error)
}
