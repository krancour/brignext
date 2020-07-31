package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

type FindSessionFn func(ctx context.Context, token string) (Session, error)

type FindUserFn func(
	ctx context.Context,
	id string,
) (sdk.User, error)

type tokenAuthFilter struct {
	findSession          FindSessionFn
	findUser             FindUserFn
	rootUserEnabled      bool
	hashedSchedulerToken string
	hashedObserverToken  string
}

func NewTokenAuthFilter(
	findSession FindSessionFn,
	findUser FindUserFn,
	rootUserEnabled bool,
	hashedSchedulerToken string,
	hashedObserverToken string,
) Filter {
	return &tokenAuthFilter{
		findSession:          findSession,
		findUser:             findUser,
		rootUserEnabled:      rootUserEnabled,
		hashedSchedulerToken: hashedSchedulerToken,
		hashedObserverToken:  hashedObserverToken,
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
				&brignext.ErrAuthentication{
					Reason: `Authorization" header is missing.`,
				},
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
				&brignext.ErrAuthentication{
					Reason: `Authorization" header is malformed.`,
				},
			)
			return
		}
		token := headerValueParts[1]

		// Is it the scheduler's token?
		if crypto.ShortSHA("", token) == t.hashedSchedulerToken {
			ctx := context.WithValue(
				r.Context(),
				principalContextKey{},
				schedulerPrincipal,
			)
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it the observer's token?
		if crypto.ShortSHA("", token) == t.hashedObserverToken {
			ctx := context.WithValue(
				r.Context(),
				principalContextKey{},
				observerPrincipal,
			)
			handle(w, r.WithContext(ctx))
			return
		}

		session, err := t.findSession(r.Context(), token)
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
				t.writeResponse(
					w,
					http.StatusUnauthorized,
					&brignext.ErrAuthentication{
						Reason: "Session not found. Please log in again.",
					},
				)
				return
			}
			log.Println(err)
			t.writeResponse(
				w,
				http.StatusInternalServerError,
				&brignext.ErrInternalServer{},
			)
			return
		}
		if session.Root && !t.rootUserEnabled {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&brignext.ErrAuthentication{
					Reason: "Supplied token was for an established root session, but " +
						"authentication using root credentials is no longer supported " +
						"by this server.",
				},
			)
			return
		}
		if session.Authenticated == nil {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&brignext.ErrAuthentication{
					Reason: "Supplied token has not been authenticated. Please log " +
						"in again.",
				},
			)
			return
		}
		if session.Expires != nil && time.Now().After(*session.Expires) {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&brignext.ErrAuthentication{
					Reason: "Supplied token has expired. Please log in again.",
				},
			)
			return
		}
		var principal Principal
		if session.Root {
			principal = rootPrincipal
		} else {
			user, err := t.findUser(r.Context(), session.UserID)
			if err != nil {
				log.Println(err)
				// There should never be an authenticated session for a user that
				// doesn't exist.
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&brignext.ErrInternalServer{},
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
