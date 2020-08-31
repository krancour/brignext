package authn

import (
	"context"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
)

type AuthorizeFn func(context.Context, ...Role) error

func AlwaysAuthorize(context.Context, ...Role) error {
	return nil
}

func NeverAuthorize(context.Context, ...Role) error {
	return &core.ErrAuthorization{}
}

// TODO: Implement this
func Authorize(context.Context, ...Role) error {
	return nil
}
