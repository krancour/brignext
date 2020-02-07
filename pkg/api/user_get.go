package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	username := mux.Vars(r)["username"]

	user, err := s.userStore.GetUserByUsername(username)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving user"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if user == nil {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Username  string    `json:"username"`
			FirstSeen time.Time `json:"firstSeen"`
		}{
			Username:  user.Username,
			FirstSeen: user.FirstSeen,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get user response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
