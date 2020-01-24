package storage

import "github.com/krancour/brignext/pkg/brignext"

type UserStore interface {
	CreateUser(username, password string) error
	GetUsers() ([]*brignext.User, error)
	GetUser(username string) (*brignext.User, error)
	GetUserByUsernameAndPassword(
		username string,
		password string,
	) (*brignext.User, error)
	GetUserByToken(token string) (*brignext.User, error)
	UpdateUserPassword(username, password string) error
	DeleteUser(username string) error
	CreateUserToken(username string) (string, error)
	DeleteUserToken(token string) error
}
