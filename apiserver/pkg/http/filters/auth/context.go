package auth

import (
	"context"

	"github.com/krancour/brignext"
)

type userContextKey struct{}

type sessionIDContextKey struct{}

func UserFromContext(ctx context.Context) (brignext.User, bool) {
	user := ctx.Value(userContextKey{})
	if user == nil {
		return brignext.User{}, false
	}
	return user.(brignext.User), true
}

func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}
