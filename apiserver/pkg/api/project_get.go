package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

func (s *server) projectGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	project, err := s.service.GetProject(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving project %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(project)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get project response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
