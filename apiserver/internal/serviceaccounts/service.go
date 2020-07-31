package serviceaccounts

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

type Service interface {
	Create(context.Context, brignext.ServiceAccount) (brignext.Token, error)
	List(context.Context) (brignext.ServiceAccountReferenceList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (brignext.Token, error)
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) Create(
	ctx context.Context,
	serviceAccount brignext.ServiceAccount,
) (brignext.Token, error) {
	now := time.Now()
	serviceAccount.Created = &now
	token := brignext.NewToken(crypto.NewToken(256))
	serviceAccount.HashedToken = crypto.ShortSHA("", token.Value)
	if err := s.store.Create(ctx, serviceAccount); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (s *service) List(
	ctx context.Context,
) (brignext.ServiceAccountReferenceList, error) {
	serviceAccountList, err := s.store.List(ctx)
	if err != nil {
		return serviceAccountList,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccountList, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.Get(ctx, id)
	if err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *service) GetByToken(
	ctx context.Context,
	token string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.GetByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return serviceAccount, errors.Wrap(
			err,
			"error retrieving service account from store by hashed token",
		)
	}
	return serviceAccount, nil
}

func (s *service) Lock(ctx context.Context, id string) error {
	if err := s.store.Lock(ctx, id); err != nil {
		return errors.Wrapf(
			err,
			"error locking service account %q in the store",
			id,
		)
	}
	return nil
}

func (s *service) Unlock(
	ctx context.Context,
	id string,
) (brignext.Token, error) {
	newToken := brignext.NewToken(crypto.NewToken(256))
	if err := s.store.Unlock(
		ctx,
		id,
		crypto.ShortSHA("", newToken.Value),
	); err != nil {
		return newToken, errors.Wrapf(
			err,
			"error unlocking service account %q in the store",
			id,
		)
	}
	return newToken, nil
}
