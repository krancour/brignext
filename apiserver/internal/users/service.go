package users

import (
	"context"
	"time"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

type Service interface {
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserReferenceList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
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
	now := time.Now()
	user.Created = &now
	if err := s.store.Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (s *service) List(
	ctx context.Context,
) (brignext.UserReferenceList, error) {
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
