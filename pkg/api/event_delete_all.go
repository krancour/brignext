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

	projectID := mux.Vars(r)["projectID"]

	deletePendingStr := r.URL.Query().Get("pending")
	var deletePending bool
	if deletePendingStr != "" {
		deletePending, _ = strconv.ParseBool(deletePendingStr) // nolint: errcheck
	}

	deleteRunningStr := r.URL.Query().Get("running")
	var deleteRunning bool
	if deleteRunningStr != "" {
		deleteRunning, _ = strconv.ParseBool(deleteRunningStr) // nolint: errcheck
	}

	_, ok, err := s.projectStore.GetProject(projectID)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	if err := s.projectStore.DeleteEventsByProjectID(
		projectID,
		storage.DeleteEventOptions{
			DeleteEventsWithPendingWorkers: deletePending,
			DeleteEventsWithRunningWorkers: deleteRunning,
		},
	); err != nil {
		log.Println(
			errors.Wrapf(err, "error deleting events for project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
