package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) eventDeleteAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["projectName"]
	forceDeleteStr := r.URL.Query().Get("force")
	var forceDelete bool
	if forceDeleteStr != "" {
		forceDelete, _ = strconv.ParseBool(forceDeleteStr) // nolint: errcheck
	}

	project, ok, err := s.projectStore.GetProjectByName(projectName)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving project %q", projectName),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	if err := s.projectStore.DeleteEventsByProjectID(
		project.ID,
		storage.DeleteEventOptions{
			DeleteEventsWithRunningWorkers: forceDelete,
		},
	); err != nil {
		log.Println(
			errors.Wrapf(err, "error deleting events for project %q", project.ID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
