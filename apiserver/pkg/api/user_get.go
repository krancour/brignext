package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	user, err := s.service.GetUser(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrUserNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving user %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
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
