package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type UsersStore interface {
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}
