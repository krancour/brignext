package storage

import (
	"context"
)

type Store interface {
	Events() EventsStore
	Projects() ProjectsStore
	ServiceAccounts() ServiceAccountsStore
	Sessions() SessionsStore
	Users() UsersStore

	DoTx(context.Context, func(context.Context) error) error
	CheckHealth(context.Context) error
}
