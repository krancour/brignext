package sessions

import (
	"context"
	"net/http"
	"strconv"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/users"
	"github.com/krancour/brignext/v2/internal/api"
	"github.com/krancour/brignext/v2/internal/api/auth"
	"github.com/krancour/brignext/v2/internal/crypto"
	errs "github.com/krancour/brignext/v2/internal/errors"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type endpoints struct {
	*api.BaseEndpoints
	oidcEnabled            bool
	rootUserEnabled        bool
	hashedRootUserPassword string
	oauth2Config           *oauth2.Config
	oidcTokenVerifier      *oidc.IDTokenVerifier
	service                Service
	usersService           users.Service
}

func NewEndpoints(
	baseEndpoints *api.BaseEndpoints,
	rootUserEnabled bool,
	hashedRootUserPassword string,
	oidcEnabled bool,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
	service Service,
	usersService users.Service,
) api.Endpoints {
	return &endpoints{
		BaseEndpoints:          baseEndpoints,
		rootUserEnabled:        rootUserEnabled,
		hashedRootUserPassword: hashedRootUserPassword,
		oauth2Config:           oauth2Config,
		oidcEnabled:            oidcEnabled,
		oidcTokenVerifier:      oidcTokenVerifier,
		service:                service,
		usersService:           usersService,
	}
}

func (e *endpoints) CheckHealth(ctx context.Context) error {
	if err := e.service.CheckHealth(ctx); err != nil {
		return err
	}
	return e.usersService.CheckHealth(ctx)
}

func (e *endpoints) Register(router *mux.Router) {
	// Create session
	router.HandleFunc(
		"/v2/sessions",
		e.create, // No filters applied to this request
	).Methods(http.MethodPost)

	// Delete session
	router.HandleFunc(
		"/v2/session",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	if e.oidcEnabled {
		// OIDC callback
		router.HandleFunc(
			"/auth/oidc/callback", // TODO: We should change this path
			e.completeAuth,        // No filters applied to this request
		).Methods(http.MethodGet)
	}
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		e.ServeRequest(
			api.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					if !e.rootUserEnabled {
						return nil, errs.NewErrNotSupported(
							"Authentication using root credentials is not supported by this " +
								"server.",
						)
					}
					username, password, ok := r.BasicAuth()
					if !ok {
						return nil, errs.NewErrBadRequest(
							"The request to create a new root session did not include a " +
								"valid basic auth header.",
						)
					}
					if username != "root" ||
						crypto.ShortSHA(username, password) != e.hashedRootUserPassword {
						return nil, errs.NewErrAuthentication(
							"Could not authenticate request using the supplied credentials.",
						)
					}
					return e.service.CreateRootSession(r.Context())
				},
				SuccessCode: http.StatusCreated,
			},
		)
		return
	}

	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				if !e.oidcEnabled {
					return nil, errs.NewErrNotSupported(
						"Authentication using OpenID Connect is not supported by this " +
							"server.",
					)
				}
				userSessionAuthDetails, err := e.service.CreateUserSession(r.Context())
				if err != nil {
					return nil, err
				}
				userSessionAuthDetails.AuthURL = e.oauth2Config.AuthCodeURL(
					userSessionAuthDetails.OAuth2State,
				)
				return userSessionAuthDetails, nil
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				sessionID := auth.SessionIDFromContext(r.Context())
				if sessionID == "" {
					return nil, errors.New(
						"error: delete session request authenticated, but no session ID " +
							"found in request context",
					)
				}
				return nil, e.service.Delete(r.Context(), sessionID)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) completeAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	oidcCode := r.URL.Query().Get("code")

	e.ServeHumanRequest(api.HumanRequest{
		W: w,
		EndpointLogic: func() (interface{}, error) {
			if oauth2State == "" || oidcCode == "" {
				return nil, errs.NewErrBadRequest(
					"The OpenID Connect authentication completion request lacked one " +
						"or both of the \"oauth2State\" and \"oidcCode\" query parameters.",
				)
			}
			session, err := e.service.GetByOAuth2State(
				r.Context(),
				oauth2State,
			)
			if err != nil {
				return nil, err
			}
			oauth2Token, err := e.oauth2Config.Exchange(r.Context(), oidcCode)
			if err != nil {
				return nil, err
			}
			rawIDToken, ok := oauth2Token.Extra("id_token").(string)
			if !ok {
				return nil, errors.New(
					"OAuth2 token, did not include an OpenID Connect identity token",
				)
			}
			idToken, err := e.oidcTokenVerifier.Verify(r.Context(), rawIDToken)
			if err != nil {
				return nil,
					errors.Wrap(err, "error verifying OpenID Connect identity token")
			}
			claims := struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{}
			if err = idToken.Claims(&claims); err != nil {
				return nil, errors.Wrap(
					err,
					"error decoding OpenID Connect identity token claims",
				)
			}
			// TODO: Push this logic down into the sessions service?
			user, err := e.usersService.Get(r.Context(), claims.Email)
			if err != nil {
				if _, ok := errors.Cause(err).(*errs.ErrNotFound); ok {
					// User wasn't found. That's ok. We'll create one.
					user = brignext.NewUser(claims.Email, claims.Name)
					if err = e.usersService.Create(r.Context(), user); err != nil {
						return nil, err
					}
				} else {
					// It was something else that went wrong when searching for the user.
					return nil, err
				}
			}
			if err = e.service.Authenticate(
				r.Context(),
				session.ID,
				user.ID,
			); err != nil {
				return nil, err
			}
			return []byte("You're now authenticated. You may resume using the CLI."),
				nil
		},
		SuccessCode: http.StatusOK,
	})
}
