package storage

import (
	"context"
)

type Store interface {
	Sessions() SessionsStore
	Users() UsersStore
	ServiceAccounts() ServiceAccountsStore
	Projects() ProjectsStore
	Events() EventsStore

	DoTx(context.Context, func(context.Context) error) error
}
