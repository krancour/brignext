package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) eventsDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	deleteAcceptedStr := r.URL.Query().Get("deleteAccepted")
	var deleteAccepted bool
	if deleteAcceptedStr != "" {
		deleteAccepted, _ = strconv.ParseBool(deleteAcceptedStr) // nolint: errcheck
	}

	deleteProcessingStr := r.URL.Query().Get("deleteProcessing")
	var deleteProcessing bool
	if deleteProcessingStr != "" {
		deleteProcessing, _ = strconv.ParseBool(deleteProcessingStr) // nolint: errcheck
	}

	if err := s.projectStore.DeleteEvents(
		storage.DeleteEventsCriteria{
			EventID:                mux.Vars(r)["id"],
			ProjectID:              mux.Vars(r)["projectID"],
			DeleteAcceptedEvents:   deleteAccepted,
			DeleteProcessingEvents: deleteProcessing,
		},
	); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting events"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
