package api

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) oidcAuthComplete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	if oauth2State == "" {
		http.Error(
			w,
			`query parameter "state" is missing from the request`,
			http.StatusBadRequest,
		)
		return
	}

	oidcCode := r.URL.Query().Get("code")
	if oidcCode == "" {
		http.Error(
			w,
			`query parameter "code" is missing from the request`,
			http.StatusBadRequest,
		)
		return
	}

	session, err := s.sessionStore.GetSessionByOAuth2State(oauth2State)
	if err != nil {
		http.Error(
			w,
			`could not locate a session matching the "state" query parameter`,
			http.StatusBadRequest,
		)
		return
	}

	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), oidcCode)
	if err != nil {
		http.Error(
			w,
			"failed to exchange code for token",
			http.StatusInternalServerError,
		)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(
			w,
			"no id_token field in oauth2 token",
			http.StatusInternalServerError,
		)
		return
	}

	idToken, err := s.oidcTokenVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "could not verify ID token", http.StatusInternalServerError)
		return
	}

	claims := struct {
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Email      string `json:"email"`
	}{}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "could not decode claims", http.StatusInternalServerError)
		return
	}

	user, err := s.userStore.GetUserByUsername(claims.Email)
	if err != nil {
		http.Error(
			w,
			"error searching for existing user",
			http.StatusInternalServerError,
		)
		return
	}

	var userID string
	if user == nil {
		if userID, err = s.userStore.CreateUser(claims.Email); err != nil {
			http.Error(
				w,
				"error creating user",
				http.StatusInternalServerError,
			)
			return
		}
	} else {
		userID = user.ID
	}

	if err := s.sessionStore.AuthenticateSession(session.ID, userID); err != nil {
		http.Error(
			w,
			"error updating session",
			http.StatusInternalServerError,
		)
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
