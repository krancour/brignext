package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) sessionCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State, token, err := s.sessionStore.CreateSession()
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

	s.writeResponse(w, http.StatusOK, responseBytes)
}
