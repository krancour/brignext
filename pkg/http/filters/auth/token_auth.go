package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/http/filter"
)

type TokenAuthFn func(token string) (*brignext.User, error)

// NewTokenAuthFilter returns an implementation of the filter.Filter interface
// that authenticates HTTP requests using a token
func NewTokenAuthFilter(tokenAuth TokenAuthFn) filter.Filter {
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
				user, err := tokenAuth(token)
				if err != nil || user == nil {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				// Success! Add the user and the token they used to the context.
				ctx := context.WithValue(r.Context(), userContextKey{}, user)
				ctx = context.WithValue(ctx, userTokenContextKey{}, token)
				handle(w, r.WithContext(ctx))
			}
		},
	)
}
