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

func (s *server) eventsCancel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]

	cancelProcessingStr := r.URL.Query().Get("cancelProcessing")
	var cancelProcessing bool
	if cancelProcessingStr != "" {
		cancelProcessing, _ =
			strconv.ParseBool(cancelProcessingStr) // nolint: errcheck
	}

	if eventID != "" {
		canceled, err := s.service.CancelEvent(
			r.Context(),
			eventID,
			cancelProcessing,
		)
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(err, "error canceling event %q", eventID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(
			struct {
				Canceled bool `json:"canceled"`
			}{
				Canceled: canceled,
			},
		)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling cancel event response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	canceled, err := s.service.CancelEventsByProject(
		r.Context(),
		projectID,
		cancelProcessing,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error canceling events for project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Canceled int64 `json:"canceled"`
		}{
			Canceled: canceled,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling cancel event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
