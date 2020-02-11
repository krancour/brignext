package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) eventList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	events, err := s.projectStore.GetEvents(storage.GetEventsCriteria{
		ProjecID: r.URL.Query().Get("projectID"),
	})
	if err != nil {
		log.Println(errors.Wrap(err, "error retrieving events"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(events)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list events response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
