package service

import (
	"context"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
)

type ServiceAccountsService interface { // nolint: golint
	Create(context.Context, brignext.ServiceAccount) (brignext.Token, error)
	List(context.Context) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (brignext.Token, error)
}

type serviceAccountsService struct {
	store storage.Store
}

func NewServiceAccountsService(store storage.Store) ServiceAccountsService {
	return &serviceAccountsService{
		store: store,
	}
}

func (s *serviceAccountsService) Create(
	ctx context.Context,
	serviceAccount brignext.ServiceAccount,
) (brignext.Token, error) {
	token := brignext.NewToken(crypto.NewToken(256))
	serviceAccount.HashedToken = crypto.ShortSHA("", token.Value)
	if err := s.store.ServiceAccounts().Create(ctx, serviceAccount); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (s *serviceAccountsService) List(
	ctx context.Context,
) (brignext.ServiceAccountList, error) {
	serviceAccountList, err := s.store.ServiceAccounts().List(ctx)
	if err != nil {
		return serviceAccountList,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccountList, nil
}

func (s *serviceAccountsService) Get(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.ServiceAccounts().Get(ctx, id)
	if err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *serviceAccountsService) GetByToken(
	ctx context.Context,
	token string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.ServiceAccounts().GetByHashedToken(
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

func (s *serviceAccountsService) Lock(ctx context.Context, id string) error {
	if err := s.store.ServiceAccounts().Lock(ctx, id); err != nil {
		return errors.Wrapf(
			err,
			"error locking service account %q in the store",
			id,
		)
	}
	return nil
}

func (s *serviceAccountsService) Unlock(
	ctx context.Context,
	id string,
) (brignext.Token, error) {
	newToken := brignext.NewToken(crypto.NewToken(256))
	if err := s.store.ServiceAccounts().Unlock(
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
