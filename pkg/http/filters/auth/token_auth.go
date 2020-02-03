package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/http/filter"
)

type FindSessionFn func(token string) (*brignext.Session, error)
type FindUserFn func(username string) (*brignext.User, error)

// NewTokenAuthFilter returns an implementation of the filter.Filter interface
// that authenticates HTTP requests using a token
func NewTokenAuthFilter(
	findSession FindSessionFn,
	findUser FindUserFn,
) filter.Filter {
	return filter.NewGenericFilter(
		func(handle http.HandlerFunc) http.HandlerFunc {
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
				session, err := findSession(token)
				if err != nil || session == nil {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				if !session.Authenticated || time.Now().After(session.ExpiresAt) {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				user, err := findUser(session.UserID)
				if err != nil || user == nil {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				// Success! Add the user and the session ID to the context.
				ctx := context.WithValue(r.Context(), userContextKey{}, user)
				ctx = context.WithValue(ctx, sessionIDContextKey{}, session.ID)
				handle(w, r.WithContext(ctx))
			}
		},
	)
}
