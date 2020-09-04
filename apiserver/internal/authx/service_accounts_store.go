package authx

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type ServiceAccountsStore interface {
	Create(context.Context, ServiceAccount) error
	List(
		context.Context,
		ServiceAccountsSelector,
		meta.ListOptions,
	) (ServiceAccountList, error)
	Get(context.Context, string) (ServiceAccount, error)
	GetByHashedToken(context.Context, string) (ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error
}
