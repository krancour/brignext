package authn

import "context"

type principalContextKey struct{}

type sessionIDContextKey struct{}

func ContextWithPrincipal(
	ctx context.Context,
	principal Principal,
) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

func ContextWithSessionID(
	ctx context.Context,
	sessionID string,
) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

func PincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(principalContextKey{}).(Principal)
}

func SessionIDFromContext(ctx context.Context) string {
	token := ctx.Value(sessionIDContextKey{})
	if token == nil {
		return ""
	}
	return token.(string)
}
