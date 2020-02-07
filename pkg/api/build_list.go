package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) buildList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["projectName"]

	var builds []brignext.Build
	var err error
	if projectName == "" {
		if builds, err = s.projectStore.GetBuilds(); err != nil {
			log.Println(
				errors.Wrap(err, "error retrieving all builds"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	} else {
		if builds, err =
			s.projectStore.GetBuildsByProjectName(projectName); err != nil {
			log.Println(
				errors.Wrap(err, "error retrieving builds for project"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	}

	responseBytes, err := json.Marshal(builds)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list builds response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
