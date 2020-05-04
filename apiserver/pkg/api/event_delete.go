package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) eventsDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]

	deletePendingStr := r.URL.Query().Get("deletePending")
	var deletePending bool
	if deletePendingStr != "" {
		deletePending, _ = strconv.ParseBool(deletePendingStr) // nolint: errcheck
	}

	deleteProcessingStr := r.URL.Query().Get("deleteProcessing")
	var deleteProcessing bool
	if deleteProcessingStr != "" {
		deleteProcessing, _ = strconv.ParseBool(deleteProcessingStr) // nolint: errcheck
	}

	if eventID != "" {
		deleted, err := s.service.DeleteEvent(
			r.Context(),
			eventID,
			deletePending,
			deleteProcessing,
		)
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(err, "error deleting event %q", eventID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(
			struct {
				Deleted bool `json:"deleted"`
			}{
				Deleted: deleted,
			},
		)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling delete event response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	deleted, err := s.service.DeleteEventsByProject(
		r.Context(),
		projectID,
		deletePending,
		deleteProcessing,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error deleting events for project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Deleted int64 `json:"deleted"`
		}{
			Deleted: deleted,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling delete event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
