package serviceaccounts

import (
	"context"

	brignext "github.com/krancour/brignext/v2/sdk"
)

type Store interface {
	Create(context.Context, brignext.ServiceAccount) error
	List(context.Context) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByHashedToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error
}
