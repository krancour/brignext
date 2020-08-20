package serviceaccounts

import (
	"context"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
)

type Store interface {
	Create(context.Context, brignext.ServiceAccount) error
	List(
		context.Context,
		brignext.ServiceAccountSelector,
		meta.ListOptions,
	) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByHashedToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error
}
