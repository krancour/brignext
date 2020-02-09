package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/krancour/brignext/pkg/brignext"

	"github.com/krancour/brignext/pkg/crypto"
	"github.com/pkg/errors"
)

func (s *server) sessionCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {

		if !s.apiServerConfig.RootUserEnabled() {
			s.writeResponse(w, http.StatusNotImplemented, responseEmptyJSON)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
			return
		}

		if username != "root" ||
			crypto.ShortSHA(username, password) !=
				s.apiServerConfig.HashedRootUserPassword() {
			s.writeResponse(w, http.StatusUnauthorized, responseEmptyJSON)
			return
		}

		_, _, token, err := s.sessionStore.CreateSession(
			brignext.Session{
				Root:          true,
				Authenticated: true,
				Expires:       time.Now().Add(10 * time.Minute),
			},
		)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error creating new root session"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(
			struct {
				Token string `json:"token"`
			}{
				Token: token,
			},
		)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling create root session response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusCreated, responseBytes)
		return
	}

	if s.oauth2Config == nil || s.oidcTokenVerifier == nil {
		s.writeResponse(w, http.StatusNotImplemented, responseEmptyJSON)
		return
	}

	_, oauth2State, token, err := s.sessionStore.CreateSession(brignext.Session{})
	if err != nil {
		log.Println(
			errors.Wrap(err, "error creating new session"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Token   string `json:"token"`
			AuthURL string `json:"authURL"`
		}{
			Token:   token,
			AuthURL: s.oauth2Config.AuthCodeURL(oauth2State),
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create session response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseBytes)
}
