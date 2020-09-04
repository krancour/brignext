package authx

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/lib/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// ServiceAccountsService is the specialized interface for managing ServiceAccounts. It's
// decoupled from underlying technology choices (e.g. data store) to keep
// business logic reusable and consistent while the underlying tech stack
// remains free to change.
type ServiceAccountsService interface {
	// Create creates a new ServiceAccount.
	Create(context.Context, ServiceAccount) (Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		ServiceAccountsSelector,
		meta.ListOptions,
	) (ServiceAccountList, error)

	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (ServiceAccount, error)
	// GetByToken retrieves a single ServiceAccount specified by token.
	GetByToken(context.Context, string) (ServiceAccount, error)

	// Lock removes access to the API for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single ServiceAccount specified by
	// its identifier. It returns a new Token.
	Unlock(context.Context, string) (Token, error)
}

type serviceAccountsService struct {
	authorize            AuthorizeFn
	serviceAccountsStore ServiceAccountsStore
}

// NewServiceAccountsService returns a specialized interface for managing ServiceAccounts.
func NewServiceAccountsService(serviceAccountsStore ServiceAccountsStore) ServiceAccountsService {
	return &serviceAccountsService{
		authorize:            Authorize,
		serviceAccountsStore: serviceAccountsStore,
	}
}

func (s *serviceAccountsService) Create(
	ctx context.Context,
	serviceAccount ServiceAccount,
) (Token, error) {
	if err := s.authorize(ctx, RoleAdmin()); err != nil {
		return Token{}, err
	}

	token := Token{
		Value: crypto.NewToken(256),
	}
	now := time.Now()
	serviceAccount.Created = &now
	serviceAccount.HashedToken = crypto.ShortSHA("", token.Value)
	if err := s.serviceAccountsStore.Create(ctx, serviceAccount); err != nil {
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
	selector ServiceAccountsSelector,
	opts meta.ListOptions,
) (ServiceAccountList, error) {
	if err := s.authorize(ctx, RoleReader()); err != nil {
		return ServiceAccountList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	serviceAccounts, err := s.serviceAccountsStore.List(ctx, selector, opts)
	if err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccounts, nil
}

func (s *serviceAccountsService) Get(
	ctx context.Context,
	id string,
) (ServiceAccount, error) {
	if err := s.authorize(ctx, RoleReader()); err != nil {
		return ServiceAccount{}, err
	}

	serviceAccount, err := s.serviceAccountsStore.Get(ctx, id)
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
) (ServiceAccount, error) {

	// No authz requirements here because this is is never invoked at the explicit
	// request of an end user; rather it is invoked only by the system itself.

	serviceAccount, err := s.serviceAccountsStore.GetByHashedToken(
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
	if err := s.authorize(ctx, RoleAdmin()); err != nil {
		return err
	}

	if err := s.serviceAccountsStore.Lock(ctx, id); err != nil {
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
) (Token, error) {
	if err := s.authorize(ctx, RoleAdmin()); err != nil {
		return Token{}, err
	}

	newToken := Token{
		Value: crypto.NewToken(256),
	}
	if err := s.serviceAccountsStore.Unlock(
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
