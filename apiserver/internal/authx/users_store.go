package authx

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
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

	GrantRole(ctx context.Context, userID string, roles ...Role) error
	RevokeRole(ctx context.Context, userID string, roles ...Role) error
}
