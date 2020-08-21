package users

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type Store interface {
	Create(context.Context, brignext.User) error
	List(
		context.Context,
		brignext.UsersSelector,
		meta.ListOptions,
	) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}
