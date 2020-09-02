package users

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
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
}

type service struct {
	authorize authx.AuthorizeFn
	store     Store
}

// NewService returns a specialized interface for managing Users.
func NewService(store Store) Service {
	return &service{
		authorize: authx.Authorize,
		store:     store,
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
