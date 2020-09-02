package serviceaccounts

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/core/projects"
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
	Create(context.Context, authn.ServiceAccount) (authn.Token, error)
	// List returns a ServiceAccountList.
	List(
		context.Context,
		authn.ServiceAccountsSelector,
		meta.ListOptions,
	) (authn.ServiceAccountList, error)

	// Get retrieves a single ServiceAccount specified by its identifier.
	Get(context.Context, string) (authn.ServiceAccount, error)
	// GetByToken retrieves a single ServiceAccount specified by token.
	GetByToken(context.Context, string) (authn.ServiceAccount, error)

	// Lock removes access to the API for a single ServiceAccount specified by its
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single ServiceAccount specified by
	// its identifier. It returns a new Token.
	Unlock(context.Context, string) (authn.Token, error)

	GrantRole(ctx context.Context, serviceAccountID string, role authn.Role) error
	RevokeRole(
		ctx context.Context,
		serviceAccountID string,
		role authn.Role,
	) error
}

type service struct {
	authorize    authn.AuthorizeFn
	store        Store
	projectStore projects.Store
}

// NewService returns a specialized interface for managing ServiceAccounts.
func NewService(store Store, projectsStore projects.Store) Service {
	return &service{
		authorize:    authn.Authorize,
		store:        store,
		projectStore: projectsStore,
	}
}

func (s *service) Create(
	ctx context.Context,
	serviceAccount authn.ServiceAccount,
) (authn.Token, error) {
	if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
		return authn.Token{}, err
	}

	token := authn.Token{
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
	selector authn.ServiceAccountsSelector,
	opts meta.ListOptions,
) (authn.ServiceAccountList, error) {
	if err := s.authorize(ctx, authn.RoleReader()); err != nil {
		return authn.ServiceAccountList{}, err
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
) (authn.ServiceAccount, error) {
	if err := s.authorize(ctx, authn.RoleReader()); err != nil {
		return authn.ServiceAccount{}, err
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
) (authn.ServiceAccount, error) {

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
	if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
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
) (authn.Token, error) {
	if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
		return authn.Token{}, err
	}

	newToken := authn.Token{
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

// TODO: There's a lot that could be DRYed up here
func (s *service) GrantRole(
	ctx context.Context,
	serviceAccountID string,
	role authn.Role,
) error {
	// There's really no such thing as a role with no scope, so if a scope isn't
	// specified, we'll assume global. Later, we'll check if global scope is even
	// allowed for the given role, but for now, this saves us from the other
	// problems created when no scope is specified.
	if role.Scope == "" {
		role.Scope = authn.RoleScopeGlobal
	}

	// The Role the current principal requires in order to grant a Role to another
	// principal depends on what role they're trying to grant, and possibly the
	// scope as well.

	switch role.Name {

	// Need to be an Admin to grant any of the following
	case authn.RoleNameAdmin:
		fallthrough
	case authn.RoleNameEventCreator:
		fallthrough
	case authn.RoleNameProjectCreator:
		fallthrough
	case authn.RoleNameReader:
		if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
			return err
		}
		// If scope for any of these isn't global, then this is a bad request
		if role.Scope != authn.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q must only be granted with a global scope.",
					role.Name,
				),
			}
		}

	// Need to be a ProjectAdmin to grant any of the following
	case authn.RoleNameProjectAdmin:
		fallthrough
	case authn.RoleNameProjectDeveloper:
		fallthrough
	case authn.RoleNameProjectUser:
		// If there is no scope specified or if scope is global, this is a bad
		// request
		if role.Scope == "" {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Scope must be specified when granting role %q.",
					role.Name,
				),
			}
		}
		if role.Scope == authn.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q may not be granted with a global scope.",
					role.Name,
				),
			}
		}

		if err := s.authorize(ctx, authn.RoleProjectAdmin(role.Scope)); err != nil {
			return err
		}

		// Make sure the project exists
		if _, err := s.projectStore.Get(ctx, role.Scope); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving project %q from store",
				role.Scope,
			)
		}
	}

	// Make sure the ServiceAccount exists
	_, err := s.store.Get(ctx, serviceAccountID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	if err := s.store.GrantRole(ctx, serviceAccountID, role); err != nil {
		return errors.Wrapf(
			err,
			"error granting role %q with scope %q to service account %q in the store",
			role.Name,
			role.Scope,
			serviceAccountID,
		)
	}

	return nil
}

// TODO: There's a lot that could be DRYed up here
func (s *service) RevokeRole(
	ctx context.Context,
	serviceAccountID string,
	role authn.Role,
) error {
	// There's really no such thing as a role with no scope, so if a scope isn't
	// specified, we'll assume global. Later, we'll check if global scope is even
	// allowed for the given role, but for now, this saves us from the other
	// problems created when no scope is specified.
	if role.Scope == "" {
		role.Scope = authn.RoleScopeGlobal
	}

	// The Role the current principal requires in order to revoke a Role belonging
	// to another principal depends on what role they're trying to revoke, and
	// possibly the scope as well.

	switch role.Name {

	// Need to be an Admin to revoke any of the following
	case authn.RoleNameAdmin:
		fallthrough
	case authn.RoleNameEventCreator:
		fallthrough
	case authn.RoleNameProjectCreator:
		fallthrough
	case authn.RoleNameReader:
		if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
			return err
		}
		// If scope for any of these isn't global, then this is a bad request
		if role.Scope != authn.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q must only be revoked with a global scope.",
					role.Name,
				),
			}
		}

	// Need to be a ProjectAdmin to revoke any of the following
	case authn.RoleNameProjectAdmin:
		fallthrough
	case authn.RoleNameProjectDeveloper:
		fallthrough
	case authn.RoleNameProjectUser:
		// If there is no scope specified or if scope is global, this is a bad
		// request
		if role.Scope == "" {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Scope must be specified when revoking role %q.",
					role.Name,
				),
			}
		}
		if role.Scope == authn.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q may not be revoked with a global scope.",
					role.Name,
				),
			}
		}

		if err := s.authorize(ctx, authn.RoleProjectAdmin(role.Scope)); err != nil {
			return err
		}

		// Make sure the project exists
		if _, err := s.projectStore.Get(ctx, role.Scope); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving project %q from store",
				role.Scope,
			)
		}
	}

	// Make sure the ServiceAccount exists
	_, err := s.store.Get(ctx, serviceAccountID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			serviceAccountID,
		)
	}

	if err := s.store.RevokeRole(ctx, serviceAccountID, role); err != nil {
		return errors.Wrapf(
			err,
			"error revoking role %q with scope %q from service account %q in the "+
				"store",
			role.Name,
			role.Scope,
			serviceAccountID,
		)
	}

	return nil
}