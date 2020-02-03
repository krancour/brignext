package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
	CreateUser(username string) (string, error)
	GetUsers() ([]*brignext.User, error)
	GetUser(id string) (*brignext.User, error)
	GetUserByUsername(username string) (*brignext.User, error)
	DeleteUser(username string) error
}
