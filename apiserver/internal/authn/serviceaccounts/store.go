package serviceaccounts

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type Store interface {
	Create(context.Context, authn.ServiceAccount) error
	List(
		context.Context,
		authn.ServiceAccountsSelector,
		meta.ListOptions,
	) (authn.ServiceAccountList, error)
	Get(context.Context, string) (authn.ServiceAccount, error)

	GetByHashedToken(context.Context, string) (authn.ServiceAccount, error)

	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error

	GrantRole(context.Context, authn.Role) error
	RevokeRole(context.Context, authn.Role) error
}
