package serviceaccounts

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type Store interface {
	Create(context.Context, brignext.ServiceAccount) error
	List(
		context.Context,
		brignext.ServiceAccountListOptions,
	) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByHashedToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error
}
