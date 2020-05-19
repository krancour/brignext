package service

import (
	"context"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
)

type UsersService interface {
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}

type usersService struct {
	store storage.Store
}

func NewUsersService(store storage.Store) UsersService {
	return &usersService{
		store: store,
	}
}

func (u *usersService) Create(ctx context.Context, user brignext.User) error {
	if err := u.store.Users().Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (u *usersService) List(ctx context.Context) (brignext.UserList, error) {
	userList, err := u.store.Users().List(ctx)
	if err != nil {
		return userList, errors.Wrap(err, "error retrieving users from store")
	}
	return userList, nil
}

func (u *usersService) Get(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user, err := u.store.Users().Get(ctx, id)
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
	return u.store.DoTx(ctx, func(ctx context.Context) error {
		var err error
		if err = u.store.Users().Lock(ctx, id); err != nil {
			return errors.Wrapf(err, "error locking user %q in store", id)
		}
		if _, err := u.store.Sessions().DeleteByUser(ctx, id); err != nil {
			return errors.Wrapf(
				err,
				"error removing sessions for user %q from store",
				id,
			)
		}
		return nil
	})
}

func (u *usersService) Unlock(ctx context.Context, id string) error {
	if err := u.store.Users().Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}
