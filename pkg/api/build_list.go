package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) buildList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["projectName"]

	var brigadeBuilds []*brigade.Build
	var err error
	if projectName == "" {
		if brigadeBuilds, err = s.projectStore.GetBuilds(); err != nil {
			log.Println(
				errors.Wrap(err, "error retrieving all builds"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	} else {
		projectID := brigade.ProjectID(projectName)
		if brigadeBuilds, err =
			s.projectStore.GetProjectBuilds(projectID); err != nil {
			log.Println(
				errors.Wrap(err, "error retrieving builds for project"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	}

	projectNamesByID := map[string]string{}
	brignextBuilds := make([]*brignext.Build, len(brigadeBuilds))
	for i, brigadeBuild := range brigadeBuilds {
		if _, ok := projectNamesByID[brigadeBuild.ProjectID]; !ok {
			project, err := s.projectStore.GetProject(brigadeBuild.ProjectID)
			if err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error retrieving project with id %q",
						brigadeBuild.ProjectID,
					),
				)
				s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
				return
			}
			if project == nil {
				log.Println(
					errors.Errorf(
						"could not find project with id %q",
						brigadeBuild.ProjectID,
					),
				)
				s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
				return
			}
			projectNamesByID[brigadeBuild.ProjectID] = project.Name
		}
		brignextBuilds[i] = brignext.BrigadeBuildToBrigNextBuild(
			brigadeBuild,
			projectNamesByID[brigadeBuild.ProjectID],
		)
	}

	responseBytes, err := json.Marshal(brignextBuilds)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list builds response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
