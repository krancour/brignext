package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) eventDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

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

	if err := s.projectStore.DeleteEvent(
		id,
		storage.DeleteEventOptions{
			DeleteEventsWithPendingWorkers: deletePending,
			DeleteEventsWithRunningWorkers: deleteRunning,
		},
	); err != nil {
		log.Println(
			errors.Wrapf(err, "error deleting event %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated workers

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
