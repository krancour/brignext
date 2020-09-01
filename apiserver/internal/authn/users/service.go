package users

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/core/projects"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// Service is the specialized interface for managing Users. It's decoupled from
// underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type Service interface {
	// Create creates a new User.
	Create(context.Context, authn.User) error
	// List returns a UserList.
	List(
		context.Context,
		authn.UsersSelector,
		meta.ListOptions,
	) (authn.UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (authn.User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error

	GrantRole(ctx context.Context, userID string, role authn.Role) error
	RevokeRole(ctx context.Context, userID string, role authn.Role) error
}

type service struct {
	authorize    authn.AuthorizeFn
	store        Store
	projectStore projects.Store
}

// NewService returns a specialized interface for managing Users.
func NewService(store Store, projectStore projects.Store) Service {
	return &service{
		authorize:    authn.Authorize,
		store:        store,
		projectStore: projectStore,
	}
}

func (s *service) Create(ctx context.Context, user authn.User) error {

	// No authz requirements here because this is is never invoked at the explicit
	// request of an end user; rather it is invoked only by the system itself.

	now := time.Now()
	user.Created = &now
	if err := s.store.Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (s *service) List(
	ctx context.Context,
	selector authn.UsersSelector,
	opts meta.ListOptions,
) (authn.UserList, error) {
	if err := s.authorize(ctx, authn.RoleReader()); err != nil {
		return authn.UserList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	users, err := s.store.List(ctx, selector, opts)
	if err != nil {
		return users, errors.Wrap(err, "error retrieving users from store")
	}
	return users, nil
}

func (s *service) Get(ctx context.Context, id string) (authn.User, error) {
	if err := s.authorize(ctx, authn.RoleReader()); err != nil {
		return authn.User{}, err
	}

	user, err := s.store.Get(ctx, id)
	if err != nil {
		return user, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, nil
}

func (s *service) Lock(ctx context.Context, id string) error {
	if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
		return err
	}

	if err := s.store.Lock(ctx, id); err != nil {
		return errors.Wrapf(err, "error locking user %q in store", id)
	}
	return nil
}

func (s *service) Unlock(ctx context.Context, id string) error {
	if err := s.authorize(ctx, authn.RoleAdmin()); err != nil {
		return err
	}

	if err := s.store.Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

// TODO: There's a lot that could be DRYed up here
func (s *service) GrantRole(
	ctx context.Context,
	userID string,
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

	// Make sure the User exists
	_, err := s.store.Get(ctx, userID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving user %q from store",
			userID,
		)
	}

	if err := s.store.GrantRole(ctx, userID, role); err != nil {
		return errors.Wrapf(
			err,
			"error granting role %q with scope %q to user %q in the store",
			role.Name,
			role.Scope,
			userID,
		)
	}

	return nil
}

// TODO: There's a lot that could be DRYed up here
func (s *service) RevokeRole(
	ctx context.Context,
	userID string,
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

	// Make sure the User exists
	_, err := s.store.Get(ctx, userID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving user %q from store",
			userID,
		)
	}

	if err := s.store.RevokeRole(ctx, userID, role); err != nil {
		return errors.Wrapf(
			err,
			"error revoking role %q with scope %q from user %q in the store",
			role.Name,
			role.Scope,
			userID,
		)
	}

	return nil
}
