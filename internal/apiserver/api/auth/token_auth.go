package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/pkg/errors"
)

type FindSessionFn func(ctx context.Context, token string) (Session, error)

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

// TODO: Access by service accounts isn't implemented yet
func (t *tokenAuthFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get("Authorization")
		if headerValue == "" {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication("\"Authorization\" header is missing"),
			)
			return
		}
		headerValueParts := strings.SplitN(
			r.Header.Get("Authorization"),
			" ",
			2,
		)
		if len(headerValueParts) != 2 || headerValueParts[0] != "Bearer" {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication("\"Authorization\" header is malformed"),
			)
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
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication("session not found"),
			)
			return
		}
		if (session.Root && !t.rootUserEnabled) ||
			session.Authenticated == nil ||
			(session.Expires != nil && time.Now().After(*session.Expires)) {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication("session not found"),
			)
			return
		}
		var principal Principal
		if session.Root {
			principal = rootPrincipal
		} else {
			user, err := t.findUser(r.Context(), session.UserID)
			if err != nil {
				t.writeResponse(
					w,
					http.StatusUnauthorized,
					brignext.NewErrAuthentication("user not found"),
				)
				return
			}
			if user.Locked != nil {
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

func (t *tokenAuthFilter) writeResponse(
	w http.ResponseWriter,
	statusCode int,
	response interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	responseBody, ok := response.([]byte)
	if !ok {
		var err error
		if responseBody, err = json.Marshal(response); err != nil {
			log.Println(errors.Wrap(err, "error marshaling response body"))
		}
	}
	if _, err := w.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}
