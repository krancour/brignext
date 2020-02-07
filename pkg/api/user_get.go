package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	username := mux.Vars(r)["username"]

	user, ok, err := s.userStore.GetUser(username)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving user %q", username),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(user)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get user response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
