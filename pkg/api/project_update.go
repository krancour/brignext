package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) projectUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["name"]
	projectID := brigade.ProjectID(projectName)

	if existingProject, err :=
		s.projectStore.GetProject(projectID); err != nil {
		log.Println(
			errors.Wrap(err, "error checking for existing project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if existingProject == nil {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of update project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	project := &brignext.Project{}
	if err := json.Unmarshal(bodyBytes, project); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of update project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if projectName != project.Name {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// TODO: We should do some kind of validation!

	brigadeProject := brignext.BrigNextProjectToBrigadeProject(project)

	if err := s.projectStore.UpdateProject(brigadeProject); err != nil {
		log.Println(
			errors.Wrap(err, "error updating project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if err := s.oldProjectStore.ReplaceProject(brigadeProject); err != nil {
		log.Println(
			errors.Wrap(err, "error updating project in old store"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	project = brignext.BrigadeProjectToBrigNextProject(brigadeProject)

	responseBytes, err := json.Marshal(project)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling update project response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
