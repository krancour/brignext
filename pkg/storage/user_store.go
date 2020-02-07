package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
	CreateUser(username string) (string, error)
	GetUsers() ([]*brignext.User, error)
	GetUser(id string) (*brignext.User, error)
	GetUserByUsername(username string) (*brignext.User, error)
	DeleteUser(id string) error
	DeleteUserByUsername(username string) error
	CreateServiceAccount(name string, description string) (string, string, error)
	GetServiceAccounts() ([]*brignext.ServiceAccount, error)
	GetServiceAccount(id string) (*brignext.ServiceAccount, error)
	GetServiceAccountByName(name string) (*brignext.ServiceAccount, error)
	DeleteServiceAccount(id string) error
	DeleteServiceAccountByName(name string) error
}
