package users

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, authx.User) error
	Count(context.Context) (int64, error)
	List(
		context.Context,
		authx.UsersSelector,
		meta.ListOptions,
	) (authx.UserList, error)
	Get(context.Context, string) (authx.User, error)

	Lock(context.Context, string) error
	Unlock(context.Context, string) error

	GrantRole(ctx context.Context, userID string, roles ...authx.Role) error
	RevokeRole(ctx context.Context, userID string, roles ...authx.Role) error
}
