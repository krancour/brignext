package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
	CreateUser(user brignext.User) error
	GetUsers() ([]brignext.User, error)
	GetUser(id string) (brignext.User, bool, error)
	LockUser(id string) error
	UnlockUser(id string) error
	CreateServiceAccount(serviceAccount brignext.ServiceAccount) (string, string, error)
	GetServiceAccounts() ([]brignext.ServiceAccount, error)
	GetServiceAccount(name string) (brignext.ServiceAccount, bool, error)
	GetServiceAccountByToken(token string) (brignext.ServiceAccount, bool, error)
	DeleteServiceAccount(name string) error
}
