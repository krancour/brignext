package core

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/meta"
)

type SecretsStore interface {
	List(ctx context.Context,
		project Project,
		opts meta.ListOptions,
	) (SecretList, error)
	Set(ctx context.Context, project Project, secret Secret) error
	Unset(ctx context.Context, project Project, key string) error
}
