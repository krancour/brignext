package api

import (
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"

	"github.com/pkg/errors"
)

func (s *server) oidcAuthComplete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	if oauth2State == "" {
		s.writeResponse(w, http.StatusBadRequest, responseOIDCAuthError)
		return
	}

	oidcCode := r.URL.Query().Get("code")
	if oidcCode == "" {
		s.writeResponse(w, http.StatusBadRequest, responseOIDCAuthError)
		return
	}

	session, ok, err := s.service.GetSessionByOAuth2State(
		r.Context(),
		oauth2State,
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving session by OAuth2 state [REDACTED]"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusBadRequest, responseOIDCAuthError)
		return
	}

	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), oidcCode)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error exchanging authorization code for OAuth2 token"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Println(
			"OAuth2 token, did not include and OpenID Connect identity token",
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	idToken, err := s.oidcTokenVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error verifying OpenID Connect identity token"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	claims := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}
	if err := idToken.Claims(&claims); err != nil {
		log.Println(
			errors.Wrap(err, "error decoding OpenID Connect identity token claims"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	user, ok, err := s.service.GetUser(r.Context(), claims.Email)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error searching for existing user %q", claims.Email),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	} else if !ok {
		user = brignext.User{
			ID:   claims.Email,
			Name: claims.Name,
		}
		if err = s.service.CreateUser(r.Context(), user); err != nil {
			log.Println(
				errors.Wrapf(err, "error creating new user %q", user.ID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
			return
		}
	}

	if _, err := s.service.AuthenticateSession(
		r.Context(),
		session.ID,
		user.ID,
	); err != nil {
		log.Println(
			errors.Wrapf(err, "error authenticating session %q", session.ID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(
		[]byte("You're now authenticated. You may resume using the CLI."),
	); err != nil {
		log.Println(
			errors.Wrap(err, "api server error: error writing response"),
		)
	}
}
