package authx

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

// UsersService is the specialized interface for managing Users. It's decoupled from
// underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type UsersService interface {
	// List returns a UserList.
	List(
		context.Context,
		UsersSelector,
		meta.ListOptions,
	) (UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (User, error)

	// Lock removes access to the API for a single User specified by their
	// identifier.
	Lock(context.Context, string) error
	// Unlock restores access to the API for a single User specified by their
	// identifier.
	Unlock(context.Context, string) error
}

type usersService struct {
	authorize  AuthorizeFn
	usersStore UsersStore
}

// NewUsersService returns a specialized interface for managing Users.
func NewUsersService(usersStore UsersStore) UsersService {
	return &usersService{
		authorize:  Authorize,
		usersStore: usersStore,
	}
}

func (u *usersService) List(
	ctx context.Context,
	selector UsersSelector,
	opts meta.ListOptions,
) (UserList, error) {
	if err := u.authorize(ctx, RoleReader()); err != nil {
		return UserList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	users, err := u.usersStore.List(ctx, selector, opts)
	if err != nil {
		return users, errors.Wrap(err, "error retrieving users from store")
	}
	return users, nil
}

func (u *usersService) Get(ctx context.Context, id string) (User, error) {
	if err := u.authorize(ctx, RoleReader()); err != nil {
		return User{}, err
	}

	user, err := u.usersStore.Get(ctx, id)
	if err != nil {
		return user, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, nil
}

func (u *usersService) Lock(ctx context.Context, id string) error {
	if err := u.authorize(ctx, RoleAdmin()); err != nil {
		return err
	}

	if err := u.usersStore.Lock(ctx, id); err != nil {
		return errors.Wrapf(err, "error locking user %q in store", id)
	}
	return nil
}

func (u *usersService) Unlock(ctx context.Context, id string) error {
	if err := u.authorize(ctx, RoleAdmin()); err != nil {
		return err
	}

	if err := u.usersStore.Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}
