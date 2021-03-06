package authx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type User struct {
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	Name            string     `json:"name" bson:"name"`
	Locked          *time.Time `json:"locked" bson:"locked"`
	UserRoles       []Role     `json:"roles,omitempty" bson:"roles,omitempty"`
}

func (u *User) Roles() []Role {
	return u.UserRoles
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "User",
			},
			Alias: (Alias)(u),
		},
	)
}

// UserList is an ordered and pageable list of Users.
type UserList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Users.
	Items []User `json:"items,omitempty"`
}

func (u UserList) MarshalJSON() ([]byte, error) {
	type Alias UserList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserList",
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersService is the specialized interface for managing Users. It's decoupled
// from underlying technology choices (e.g. data store) to keep business logic
// reusable and consistent while the underlying tech stack remains free to
// change.
type UsersService interface {
	// List returns a UserList.
	List(context.Context, meta.ListOptions) (UserList, error)
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
	opts meta.ListOptions,
) (UserList, error) {
	if err := u.authorize(ctx, RoleReader()); err != nil {
		return UserList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	users, err := u.usersStore.List(ctx, opts)
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

type UsersStore interface {
	Create(context.Context, User) error
	Count(context.Context) (int64, error)
	List(context.Context, meta.ListOptions) (UserList, error)
	Get(context.Context, string) (User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error
}
