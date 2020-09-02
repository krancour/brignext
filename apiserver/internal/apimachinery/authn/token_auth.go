package authn

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/crypto"
	"github.com/pkg/errors"
)

type FindSessionFn func(
	ctx context.Context,
	token string,
) (authx.Session, error)

type FindEventFn func(ctx context.Context, token string) (core.Event, error)

type FindUserFn func(ctx context.Context, id string) (authx.User, error)

type FindServiceAccountFn func(
	ctx context.Context,
	token string,
) (authx.ServiceAccount, error)

type tokenAuthFilter struct {
	findSession          FindSessionFn
	findEvent            FindEventFn
	findUser             FindUserFn
	findServiceAccount   FindServiceAccountFn
	rootUserEnabled      bool
	hashedSchedulerToken string
	hashedObserverToken  string
}

func NewTokenAuthFilter(
	findSession FindSessionFn,
	findEvent FindEventFn,
	findUser FindUserFn,
	findServiceAccount FindServiceAccountFn,
	rootUserEnabled bool,
	hashedSchedulerToken string,
	hashedObserverToken string,
) apimachinery.Filter {
	return &tokenAuthFilter{
		findSession:          findSession,
		findEvent:            findEvent,
		findUser:             findUser,
		findServiceAccount:   findServiceAccount,
		rootUserEnabled:      rootUserEnabled,
		hashedSchedulerToken: hashedSchedulerToken,
		hashedObserverToken:  hashedObserverToken,
	}
}

func (t *tokenAuthFilter) Decorate(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get("Authorization")
		if headerValue == "" {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&core.ErrAuthentication{
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
				&core.ErrAuthentication{
					Reason: `Authorization" header is malformed.`,
				},
			)
			return
		}
		token := headerValueParts[1]

		// Is it the Scheduler's token?
		if crypto.ShortSHA("", token) == t.hashedSchedulerToken {
			ctx := authx.ContextWithPrincipal(r.Context(), authx.Scheduler)
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it the Observer's token?
		if crypto.ShortSHA("", token) == t.hashedObserverToken {
			ctx := authx.ContextWithPrincipal(r.Context(), authx.Observer)
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it a Worker's token?
		if event, err := t.findEvent(r.Context(), token); err != nil {
			if _, ok := errors.Cause(err).(*core.ErrNotFound); !ok {
				log.Println(err)
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&core.ErrInternalServer{},
				)
				return
			}
		} else {
			ctx := authx.ContextWithPrincipal(r.Context(), authx.Worker(event.ID))
			handle(w, r.WithContext(ctx))
			return
		}

		// Is it a ServiceAccount's token?
		if serviceAccount, err :=
			t.findServiceAccount(r.Context(), token); err != nil {
			if _, ok := errors.Cause(err).(*core.ErrNotFound); !ok {
				log.Println(err)
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&core.ErrInternalServer{},
				)
				return
			}
		} else {
			if serviceAccount.Locked != nil {
				http.Error(w, "{}", http.StatusForbidden)
				return
			}
			ctx := authx.ContextWithPrincipal(r.Context(), &serviceAccount)
			handle(w, r.WithContext(ctx))
			return
		}

		session, err := t.findSession(r.Context(), token)
		if err != nil {
			if _, ok := errors.Cause(err).(*core.ErrNotFound); ok {
				t.writeResponse(
					w,
					http.StatusUnauthorized,
					&core.ErrAuthentication{
						Reason: "Session not found. Please log in again.",
					},
				)
				return
			}
			log.Println(err)
			t.writeResponse(
				w,
				http.StatusInternalServerError,
				&core.ErrInternalServer{},
			)
			return
		}
		if session.Root && !t.rootUserEnabled {
			t.writeResponse(
				w,
				http.StatusUnauthorized,
				&core.ErrAuthentication{
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
				&core.ErrAuthentication{
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
				&core.ErrAuthentication{
					Reason: "Supplied token has expired. Please log in again.",
				},
			)
			return
		}
		var principal authx.Principal
		if session.Root {
			principal = authx.Root
		} else {
			user, err := t.findUser(r.Context(), session.UserID)
			if err != nil {
				log.Println(err)
				// There should never be an authenticated session for a user that
				// doesn't exist.
				t.writeResponse(
					w,
					http.StatusInternalServerError,
					&core.ErrInternalServer{},
				)
				return
			}
			if user.Locked != nil {
				http.Error(w, "{}", http.StatusForbidden)
				return
			}
			principal = &user
		}

		// Success! Add the user and the session ID to the context.
		ctx := authx.ContextWithPrincipal(r.Context(), principal)
		ctx = authx.ContextWithSessionID(ctx, session.ID)
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
