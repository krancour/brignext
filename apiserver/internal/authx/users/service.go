package users

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
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
	// List returns a UserList.
	List(
		context.Context,
		authx.UsersSelector,
		meta.ListOptions,
	) (authx.UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (authx.User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error

	GrantRole(ctx context.Context, userID string, role authx.Role) error
	RevokeRole(ctx context.Context, userID string, role authx.Role) error
}

type service struct {
	authorize     authx.AuthorizeFn
	store         Store
	projectsStore projects.Store
}

// NewService returns a specialized interface for managing Users.
func NewService(store Store, projectsStore projects.Store) Service {
	return &service{
		authorize:     authx.Authorize,
		store:         store,
		projectsStore: projectsStore,
	}
}

func (s *service) List(
	ctx context.Context,
	selector authx.UsersSelector,
	opts meta.ListOptions,
) (authx.UserList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return authx.UserList{}, err
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

func (s *service) Get(ctx context.Context, id string) (authx.User, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return authx.User{}, err
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
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
		return err
	}

	if err := s.store.Lock(ctx, id); err != nil {
		return errors.Wrapf(err, "error locking user %q in store", id)
	}
	return nil
}

func (s *service) Unlock(ctx context.Context, id string) error {
	if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
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
	role authx.Role,
) error {
	// There's really no such thing as a role with no scope, so if a scope isn't
	// specified, we'll assume global. Later, we'll check if global scope is even
	// allowed for the given role, but for now, this saves us from the other
	// problems created when no scope is specified.
	if role.Scope == "" {
		role.Scope = authx.RoleScopeGlobal
	}

	// The Role the current principal requires in order to grant a Role to another
	// principal depends on what role they're trying to grant, and possibly the
	// scope as well.

	switch role.Name {

	// Need to be an Admin to grant any of the following
	case authx.RoleNameAdmin:
		fallthrough
	case authx.RoleNameEventCreator:
		fallthrough
	case authx.RoleNameProjectCreator:
		fallthrough
	case authx.RoleNameReader:
		if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
			return err
		}
		// If scope for any of these isn't global, then this is a bad request
		if role.Scope != authx.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q must only be granted with a global scope.",
					role.Name,
				),
			}
		}

	// Need to be a ProjectAdmin to grant any of the following
	case authx.RoleNameProjectAdmin:
		fallthrough
	case authx.RoleNameProjectDeveloper:
		fallthrough
	case authx.RoleNameProjectUser:
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
		if role.Scope == authx.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q may not be granted with a global scope.",
					role.Name,
				),
			}
		}

		if err := s.authorize(ctx, authx.RoleProjectAdmin(role.Scope)); err != nil {
			return err
		}

		// Make sure the project exists
		if _, err := s.projectsStore.Get(ctx, role.Scope); err != nil {
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
	role authx.Role,
) error {
	// There's really no such thing as a role with no scope, so if a scope isn't
	// specified, we'll assume global. Later, we'll check if global scope is even
	// allowed for the given role, but for now, this saves us from the other
	// problems created when no scope is specified.
	if role.Scope == "" {
		role.Scope = authx.RoleScopeGlobal
	}

	// The Role the current principal requires in order to revoke a Role belonging
	// to another principal depends on what role they're trying to revoke, and
	// possibly the scope as well.

	switch role.Name {

	// Need to be an Admin to revoke any of the following
	case authx.RoleNameAdmin:
		fallthrough
	case authx.RoleNameEventCreator:
		fallthrough
	case authx.RoleNameProjectCreator:
		fallthrough
	case authx.RoleNameReader:
		if err := s.authorize(ctx, authx.RoleAdmin()); err != nil {
			return err
		}
		// If scope for any of these isn't global, then this is a bad request
		if role.Scope != authx.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q must only be revoked with a global scope.",
					role.Name,
				),
			}
		}

	// Need to be a ProjectAdmin to revoke any of the following
	case authx.RoleNameProjectAdmin:
		fallthrough
	case authx.RoleNameProjectDeveloper:
		fallthrough
	case authx.RoleNameProjectUser:
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
		if role.Scope == authx.RoleScopeGlobal {
			return &core.ErrBadRequest{
				Reason: fmt.Sprintf(
					"Role %q may not be revoked with a global scope.",
					role.Name,
				),
			}
		}

		if err := s.authorize(ctx, authx.RoleProjectAdmin(role.Scope)); err != nil {
			return err
		}

		// Make sure the project exists
		if _, err := s.projectsStore.Get(ctx, role.Scope); err != nil {
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
