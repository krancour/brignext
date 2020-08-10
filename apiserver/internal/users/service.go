package users

import (
	"context"
	"time"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

// Service is the specialized interface for managing Users. It's decoupled from
// underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type Service interface {
	// Create creates a new User.
	Create(context.Context, brignext.User) error
	// List returns a UserList.
	List(context.Context, brignext.UserListOptions) (brignext.UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (brignext.User, error)
	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error
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

func (s *service) Create(ctx context.Context, user brignext.User) error {
	now := time.Now()
	user.Created = &now
	if err := s.store.Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (s *service) List(
	ctx context.Context,
	opts brignext.UserListOptions,
) (brignext.UserList, error) {
	userList, err := s.store.List(ctx, opts)
	if err != nil {
		return userList, errors.Wrap(err, "error retrieving users from store")
	}
	return userList, nil
}

func (s *service) Get(ctx context.Context, id string) (brignext.User, error) {
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
