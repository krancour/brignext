package auth

import (
	"context"
)

type principalContextKey struct{}

type sessionIDContextKey struct{}

func PincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(principalContextKey{})
	// user := ctx.Value(userContextKey{})
	// if user == nil {
	// 	return brignext.User{}, false
	// }
	// return user.(brignext.User), true
}

func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}
