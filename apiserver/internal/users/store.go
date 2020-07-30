package users

import (
	"context"

	brignext "github.com/krancour/brignext/v2/sdk"
)

type Store interface {
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}
