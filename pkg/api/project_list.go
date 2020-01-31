package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) projectList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projs, err := s.projectStore.GetProjects()
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all projects"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	brignextProjects := make([]*brignext.Project, len(projs))
	for i, project := range projs {
		brignextProjects[i] = brignext.BrigadeProjectToBrigNextProject(project)
	}

	responseBytes, err := json.Marshal(brignextProjects)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list projects response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
