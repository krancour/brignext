package api

import (
	"log"
	"net/http"
	"time"

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

	session, ok, err := s.sessionStore.GetSessionByOAuth2State(oauth2State)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), oidcCode)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	idToken, err := s.oidcTokenVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	claims := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}
	if err := idToken.Claims(&claims); err != nil {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}

	user, ok, err := s.userStore.GetUser(claims.Email)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
		return
	}
	if !ok {
		user = brignext.User{
			Username:  claims.Email,
			Name:      claims.Name,
			FirstSeen: time.Now(),
		}
		if s.userStore.CreateUser(user); err != nil {
			s.writeResponse(w, http.StatusInternalServerError, responseOIDCAuthError)
			return
		}
	}

	if err :=
		s.sessionStore.AuthenticateSession(session.ID, user.Username); err != nil {
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
