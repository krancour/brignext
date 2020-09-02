package serviceaccounts

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, authx.ServiceAccount) error
	List(
		context.Context,
		authx.ServiceAccountsSelector,
		meta.ListOptions,
	) (authx.ServiceAccountList, error)
	Get(context.Context, string) (authx.ServiceAccount, error)

	GetByHashedToken(context.Context, string) (authx.ServiceAccount, error)

	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error

	GrantRole(ctx context.Context, serviceAccountID string, role authx.Role) error
	RevokeRole(
		ctx context.Context,
		serviceAccountID string,
		role authx.Role,
	) error
}
