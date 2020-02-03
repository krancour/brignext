package auth

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
)

type userContextKey struct{}

type sessionIDContextKey struct{}

func UserFromContext(ctx context.Context) *brignext.User {
	user := ctx.Value(userContextKey{})
	if user == nil {
		return nil
	}
	return user.(*brignext.User)
}

func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}
