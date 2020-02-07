package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) buildGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	buildID := mux.Vars(r)["id"]

	build, err := s.projectStore.GetBuild(buildID)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving build"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if build == nil {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	project, err := s.projectStore.GetProject(build.ProjectID)
	if err != nil {
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving project with id %q",
				build.ProjectID,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	brignextBuild := brignext.BrigadeBuildToBrigNextBuild(build, project.Name)

	responseBytes, err := json.Marshal(brignextBuild)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get build response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
