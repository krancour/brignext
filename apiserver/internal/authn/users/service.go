package users

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
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

	GrantRole(context.Context, authn.Role) error
	RevokeRole(context.Context, authn.Role) error
}

type service struct {
	store Store
}

// NewService returns a specialized interface for managing Users.
func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) Create(ctx context.Context, user authn.User) error {
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
	if err := s.store.Lock(ctx, id); err != nil {
		return errors.Wrapf(err, "error locking user %q in store", id)
	}
	return nil
}

func (s *service) Unlock(ctx context.Context, id string) error {
	if err := s.store.Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

// TODO: Implement this
func (s *service) GrantRole(context.Context, authn.Role) error {
	return nil
}

// TODO: Implement this
func (s *service) RevokeRole(context.Context, authn.Role) error {
	return nil
}
