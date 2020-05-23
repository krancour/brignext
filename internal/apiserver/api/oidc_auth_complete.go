package api

import (
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// TODO: Figure out where to move all of this to
type oidcEndpoints struct {
	*baseEndpoints
	oauth2Config      *oauth2.Config
	oidcTokenVerifier *oidc.IDTokenVerifier
	service           service.Service
}

func (o *oidcEndpoints) register(router *mux.Router) {
	// OIDC callback
	router.HandleFunc(
		"/auth/oidc/callback",
		o.completeAuth, // No filters applied to this request
	).Methods(http.MethodGet)
}

func (o *oidcEndpoints) completeAuth(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	oidcCode := r.URL.Query().Get("code")

	o.serveHumanRequest(humanRequest{
		w: w,
		endpointLogic: func() (interface{}, error) {
			if oauth2State == "" || oidcCode == "" {
				return nil, brignext.NewErrBadRequest(
					"The OpenID Connect authentication completion request lacked one " +
						"or both of the \"oauth2State\" and \"oidcCode\" query parameters.",
				)
			}
			session, err := o.service.Sessions().GetByOAuth2State(
				r.Context(),
				oauth2State,
			)
			if err != nil {
				return nil, err
			}
			oauth2Token, err := o.oauth2Config.Exchange(r.Context(), oidcCode)
			if err != nil {
				return nil, err
			}
			rawIDToken, ok := oauth2Token.Extra("id_token").(string)
			if !ok {
				return nil, errors.New(
					"OAuth2 token, did not include an OpenID Connect identity token",
				)
			}
			idToken, err := o.oidcTokenVerifier.Verify(r.Context(), rawIDToken)
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
			user, err := o.service.Users().Get(r.Context(), claims.Email)
			if err != nil {
				if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
					// User wasn't found. That's ok. We'll create one.
					user = brignext.NewUser(claims.Email, claims.Name)
					if err = o.service.Users().Create(r.Context(), user); err != nil {
						return nil, err
					}
				} else {
					// It was something else that went wrong when searching for the user.
					return nil, err
				}
			}
			if err = o.service.Sessions().Authenticate(
				r.Context(),
				session.ID,
				user.ID,
			); err != nil {
				return nil, err
			}
			return []byte("You're now authenticated. You may resume using the CLI."),
				nil
		},
		successCode: http.StatusOK,
	})
}
