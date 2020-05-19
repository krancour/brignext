package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) eventList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	var eventList brignext.EventList
	var err error
	if projectID := r.URL.Query().Get("projectID"); projectID != "" {
		eventList, err = s.service.GetEventsByProject(r.Context(), projectID)
	} else {
		eventList, err = s.service.GetEvents(r.Context())
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error retrieving events"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(eventList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list events response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}