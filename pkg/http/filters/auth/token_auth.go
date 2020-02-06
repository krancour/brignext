package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
)

type FindSessionFn func(token string) (*brignext.Session, error)
type FindUserFn func(username string) (*brignext.User, error)

type tokenAuthFilter struct {
	findSession     FindSessionFn
	findUser        FindUserFn
	rootUserEnabled bool
}

func NewTokenAuthFilter(
	findSession FindSessionFn,
	findUser FindUserFn,
	rootUserEnabled bool,
) AuthFilter {
	return &tokenAuthFilter{
		findSession:     findSession,
		findUser:        findUser,
		rootUserEnabled: rootUserEnabled,
	}
}

func (t *tokenAuthFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get("Authorization")
		if headerValue == "" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		headerValueTokens := strings.SplitN(
			r.Header.Get("Authorization"),
			" ",
			2,
		)
		if len(headerValueTokens) != 2 || headerValueTokens[0] != "Bearer" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		token := headerValueTokens[1]
		session, err := t.findSession(token)
		if err != nil || session == nil {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		if (session.Root && !t.rootUserEnabled) ||
			!session.Authenticated ||
			time.Now().After(session.ExpiresAt) {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		var user *brignext.User
		if session.Root {
			user = &brignext.User{
				Username: "root",
			}
		} else {
			if user, err = t.findUser(session.UserID); err != nil || user == nil {
				http.Error(w, "{}", http.StatusUnauthorized)
				return
			}
		}
		// Success! Add the user and the session ID to the context.
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		ctx = context.WithValue(ctx, sessionIDContextKey{}, session.ID)
		handle(w, r.WithContext(ctx))
	}
}
