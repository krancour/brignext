package auth

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
)

type userContextKey struct{}

type userTokenContextKey struct{}

func UserFromContext(ctx context.Context) *brignext.User {
	user := ctx.Value(userContextKey{})
	if user == nil {
		return nil
	}
	return user.(*brignext.User)
}

func UserTokenFromContext(ctx context.Context) string {
	token := ctx.Value(userTokenContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}
