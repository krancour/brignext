package authx

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
)

type UsersStore interface {
	Create(context.Context, User) error
	Count(context.Context) (int64, error)
	List(
		context.Context,
		UsersSelector,
		meta.ListOptions,
	) (UserList, error)
	Get(context.Context, string) (User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}