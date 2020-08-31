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
	authorize authn.AuthorizeFn
	store     Store
}

// NewService returns a specialized interface for managing Users.
func NewService(store Store) Service {
	return &service{
		authorize: authn.Authorize,
		store:     store,
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

// TODO: Finish implementing this
func (s *service) GrantRole(ctx context.Context, role authn.Role) error {
	// TODO: Need to look closely at what is being granted to decide if the
	// principal is authorized to grant that.
	return nil
}

// TODO: Finish implementing this
func (s *service) RevokeRole(ctx context.Context, role authn.Role) error {
	// TODO: Need to look closely at what is being revoked to decide if the
	// principal is authorized to revoke that.
	return nil
}
