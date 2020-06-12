package users

import (
	"context"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

type Service interface {
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error

	CheckHealth(context.Context) error
}

type service struct {
	store Store
}

func NewService(store Store) Service {
	return &service{
		store: store,
	}
}

func (s *service) Create(ctx context.Context, user brignext.User) error {
	if err := s.store.Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (s *service) List(ctx context.Context) (brignext.UserList, error) {
	userList, err := s.store.List(ctx)
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
	return s.store.DoTx(ctx, func(ctx context.Context) error {
		var err error
		if err = s.store.Lock(ctx, id); err != nil {
			return errors.Wrapf(err, "error locking user %q in store", id)
		}
		// TODO: Cascade this delete somehow
		// if _, err := u.store.Sessions().DeleteByUser(ctx, id); err != nil {
		// 	return errors.Wrapf(
		// 		err,
		// 		"error removing sessions for user %q from store",
		// 		id,
		// 	)
		// }
		return nil
	})
}

func (s *service) Unlock(ctx context.Context, id string) error {
	if err := s.store.Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

func (s *service) CheckHealth(ctx context.Context) error {
	return s.store.CheckHealth(ctx)
}
