package auth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authHeader = "authorization"

type userContextKey struct{}

type userTokenContextKey struct{}

// Authenticator exposes a function for authenticating requests.
type Authenticator struct {
	userStore storage.UserStore
}

func NewAuthenticator(userStore storage.UserStore) *Authenticator {
	return &Authenticator{
		userStore: userStore,
	}
}

func (a *Authenticator) Authenticate(
	ctx context.Context,
) (context.Context, error) {
	sts := grpc.ServerTransportStreamFromContext(ctx)
	if sts == nil {
		return ctx, status.Error(
			codes.Unauthenticated,
			"no server transport stream associated with context",
		)
	}
	switch sts.Method() {
	case "/users.Users/Register":
		// Bypass authentication for registration
		return ctx, nil
	case "/users.Users/Login":
		// User basic auth for login
		return a.basicAuth(ctx)
	default:
		// All other cases use token auth
		return a.tokenAuth(ctx)
	}
}

func (a Authenticator) basicAuth(
	ctx context.Context,
) (context.Context, error) {
	auth, err := extractAuthHeader(ctx)
	if err != nil {
		return ctx, err
	}
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return ctx, status.Error(
			codes.Unauthenticated,
			`missing "Basic " prefix in "Authorization" header`,
		)
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, `invalid base64 in header`)
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return ctx, status.Error(codes.Unauthenticated, `invalid basic auth format`)
	}
	username, password := cs[:s], cs[s+1:]

	user, err := a.userStore.GetUserByUsernameAndPassword(username, password)
	if err != nil {
		return ctx, status.Errorf(
			codes.Unauthenticated,
			"error retrieving user %q",
			username,
		)
	}
	if user == nil {
		return ctx, status.Error(
			codes.Unauthenticated, "invalid username or password",
		)
	}

	ctx = context.WithValue(ctx, userContextKey{}, user)

	return purgeAuthHeader(ctx), nil
}

func (a Authenticator) tokenAuth(ctx context.Context) (context.Context, error) {
	auth, err := extractAuthHeader(ctx)
	if err != nil {
		return ctx, err
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ctx, status.Error(
			codes.Unauthenticated,
			`missing "Bearer " prefix in "Authorization" header`,
		)
	}
	token := strings.TrimPrefix(auth, prefix)

	user, err := a.userStore.GetUserByToken(token)
	if err != nil {
		return ctx, status.Error(
			codes.Unauthenticated,
			"error retrieving user by token [REDACTED]",
		)
	}
	if user == nil {
		return ctx, status.Error(
			codes.Unauthenticated, "invalid token",
		)
	}

	ctx = context.WithValue(ctx, userContextKey{}, user)
	ctx = context.WithValue(ctx, userTokenContextKey{}, token)

	return purgeAuthHeader(ctx), nil
}

func extractAuthHeader(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers in request")
	}

	authHeaders, ok := md[authHeader]
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no header in request")
	}

	if len(authHeaders) != 1 {
		return "", status.Error(
			codes.Unauthenticated,
			"more than 1 header in request",
		)
	}

	return authHeaders[0], nil
}

func purgeAuthHeader(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	mdCopy := md.Copy()
	mdCopy[authHeader] = nil
	return metadata.NewIncomingContext(ctx, mdCopy)
}

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
