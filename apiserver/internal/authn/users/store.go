package users

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, authn.User) error
	Count(context.Context) (int64, error)
	List(
		context.Context,
		authn.UsersSelector,
		meta.ListOptions,
	) (authn.UserList, error)
	Get(context.Context, string) (authn.User, error)

	Lock(context.Context, string) error
	Unlock(context.Context, string) error

	GrantRole(ctx context.Context, userID string, role authn.Role) error
	RevokeRole(ctx context.Context, userID string, role authn.Role) error
}
