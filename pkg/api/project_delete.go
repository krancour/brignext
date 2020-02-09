package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) projectDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	if err := s.projectStore.DeleteProjectByName(name); err != nil {
		log.Println(
			errors.Wrapf(err, "error deleting project %q", name),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated events and jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
