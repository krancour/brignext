package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) projectGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	project, ok, err := s.projectStore.GetProjectByName(name)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving project %q", name),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
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
