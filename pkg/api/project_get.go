package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) projectGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["name"]
	projectID := brigade.ProjectID(projectName)

	project, err := s.projectStore.GetProject(projectID)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if project == nil {
		s.writeResponse(w, http.StatusOK, responseEmptyJSON)
		return
	}

	brignextProject := brignext.BrigadeProjectToBrigNextProject(project)

	responseBytes, err := json.Marshal(brignextProject)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get project response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
