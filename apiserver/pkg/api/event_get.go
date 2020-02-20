package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) eventGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	event, err := s.service.GetEvent(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving event %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(event)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
