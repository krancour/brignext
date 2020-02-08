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

	events, err := s.projectStore.GetEventsByProjectName(projectName)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving events for project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	for _, event := range events {
		if err := s.projectStore.DeleteEvent(
			event.ID,
			storage.DeleteEventOptions{
				DeleteEventsWithRunningWorkers: forceDelete,
			},
		); err != nil {
			log.Println(
				errors.Wrap(err, "error deleting event"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
