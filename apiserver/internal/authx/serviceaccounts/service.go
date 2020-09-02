package serviceaccounts

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// Service is the specialized interface for managing ServiceAccounts. It's
// decoupled from underlying technology choices (e.g. data store) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type Service interface {
	// Create creates a new ServiceAccount.
	Create(context.Context, authx.ServiceAccount) (authx.Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		authx.ServiceAccountsSelector,
		meta.ListOptions,
	) (authx.ServiceAccountList, error)

	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (authx.ServiceAccount, error)
	// GetByToken retrieves a single ServiceAccount specified by token.
	GetByToken(context.Context, string) (authx.ServiceAccount, error)

	// Lock removes access to the API for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single ServiceAccount specified by
	// its identifier. It returns a new Token.
	Unlock(context.Context, string) (authx.Token, error)
}

type service struct {
	authorize authx.AuthorizeFn
	store     Store
}

// NewService returns a specialized interface for managing ServiceAccounts.
func NewService(store Store) Service {
	return &service{
		authorize: authx.Authorize,
		store:     store,
	}
}

func (s *service) Create(
	ctx context.Context,
	serviceAccount authx.ServiceAccount,
) (authx.Token, error) {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return authx.Token{}, err
	}

	token := authx.Token{
		Value: crypto.NewToken(256),
	}
	now := time.Now()
	serviceAccount.Created = &now
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
	selector authx.ServiceAccountsSelector,
	opts meta.ListOptions,
) (authx.ServiceAccountList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return authx.ServiceAccountList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	serviceAccounts, err := s.store.List(ctx, selector, opts)
	if err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccounts, nil
}

func (s *service) Get(
	ctx context.Context,
	id string,
) (authx.ServiceAccount, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return authx.ServiceAccount{}, err
	}

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
) (authx.ServiceAccount, error) {

	// No authz requirements here because this is is never invoked at the explicit
	// request of an end user; rather it is invoked only by the system itself.

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
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

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
) (authx.Token, error) {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return authx.Token{}, err
	}

	newToken := authx.Token{
		Value: crypto.NewToken(256),
	}
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
