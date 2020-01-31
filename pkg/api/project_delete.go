package api

import (
	"log"
	"net/http"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) projectDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["name"]
	projectID := brigade.ProjectID(projectName)

	if err := s.projectStore.DeleteProject(projectID); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if err := s.oldProjectStore.DeleteProject(projectID); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting project from old store"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated builds and jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
