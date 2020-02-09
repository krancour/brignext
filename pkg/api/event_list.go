package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) eventList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]

	var events []brignext.Event
	var err error
	if projectID == "" {
		if events, err = s.projectStore.GetEvents(); err != nil {
			log.Println(
				errors.Wrap(err, "error retrieving all events"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	} else {
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

		if events, err =
			s.projectStore.GetEventsByProjectID(projectID); err != nil {
			log.Println(
				errors.Wrapf(err, "error retrieving events for project %q", projectID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
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