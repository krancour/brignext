package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/http/filters/auth"
	"github.com/pkg/errors"
)

func (s *server) tokenCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	user := auth.UserFromContext(r.Context())
	if user == nil {
		log.Println(
			"error: create token request authenticated, but no user found in " +
				"request context",
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	token, err := s.userStore.CreateUserToken(user.Username)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error creating new user token"),
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
			errors.Wrap(err, "error marshaling create token response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
