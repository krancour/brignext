package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
)

type FindSessionFn func(
	ctx context.Context,
	token string,
) (brignext.Session, error)

type FindUserFn func(
	ctx context.Context,
	id string,
) (brignext.User, error)

type tokenAuthFilter struct {
	findSession           FindSessionFn
	findUser              FindUserFn
	rootUserEnabled       bool
	hashedControllerToken string
}

func NewTokenAuthFilter(
	findSession FindSessionFn,
	findUser FindUserFn,
	rootUserEnabled bool,
	hashedControllerToken string,
) Filter {
	return &tokenAuthFilter{
		findSession:           findSession,
		findUser:              findUser,
		rootUserEnabled:       rootUserEnabled,
		hashedControllerToken: hashedControllerToken,
	}
}

func (t *tokenAuthFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get("Authorization")
		if headerValue == "" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		headerValueParts := strings.SplitN(
			r.Header.Get("Authorization"),
			" ",
			2,
		)
		if len(headerValueParts) != 2 || headerValueParts[0] != "Bearer" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		token := headerValueParts[1]

		// Is it the controller's token?
		if crypto.ShortSHA("", token) == t.hashedControllerToken {
			ctx := context.WithValue(
				r.Context(),
				principalContextKey{},
				controllerPrincipal,
			)
			handle(w, r.WithContext(ctx))
			return
		}

		session, err := t.findSession(r.Context(), token)
		if err != nil {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		if (session.Root && !t.rootUserEnabled) ||
			!session.Authenticated ||
			time.Now().After(session.Expires) {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		var principal Principal
		if session.Root {
			principal = rootPrincipal
		} else {
			user, err := t.findUser(r.Context(), session.UserID)
			if err != nil {
				http.Error(w, "{}", http.StatusUnauthorized)
				return
			}
			if user.Locked {
				http.Error(w, "{}", http.StatusForbidden)
				return
			}
			principal = user
		}

		// Success! Add the user and the session ID to the context.
		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		ctx = context.WithValue(ctx, sessionIDContextKey{}, session.ID)
		handle(w, r.WithContext(ctx))
	}
}
