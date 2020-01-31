package auth

import (
	"context"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/http/filter"
)

type UsernamePasswordAuthFn func(
	username string,
	password string,
) (*brignext.User, error)

// NewBasicAuthFilter returns an implementation of the filter.Filter interface
// that authenticates HTTP requests using Basic Auth
func NewBasicAuthFilter(
	usernamePasswordAuth UsernamePasswordAuthFn,
) filter.Filter {
	return filter.NewGenericFilter(
		func(handle http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				if !ok {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				user, err := usernamePasswordAuth(username, password)
				if err != nil || user == nil {
					http.Error(w, "{}", http.StatusUnauthorized)
					return
				}
				// Success! Add the user to the context.
				ctx := context.WithValue(r.Context(), userContextKey{}, user)
				handle(w, r.WithContext(ctx))
			}
		},
	)
}
