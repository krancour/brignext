package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) eventsDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]

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

	var err error
	if eventID != "" {
		_, err = s.service.DeleteEvent(
			r.Context(),
			eventID,
			deleteAccepted,
			deleteProcessing,
		)
	} else {
		_, err = s.service.DeleteEventsByProject(
			r.Context(),
			projectID,
			deleteAccepted,
			deleteProcessing,
		)
	}
	if err != nil {
		log.Println(
			errors.Wrap(err, "error deleting events"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: We should respond with a count of how many were deleted
	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
