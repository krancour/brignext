package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
)

func (s *server) projectCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of create project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	project := &brignext.Project{}
	if err := json.Unmarshal(bodyBytes, project); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// TODO: We should do some kind of validation!

	if existingProject, err :=
		s.projectStore.GetProject(project.Name); err != nil {
		log.Println(
			errors.Wrap(err, "error checking for existing project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if existingProject != nil {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	brigadeProject := brignext.BrigNextProjectToBrigadeProject(project)

	if err := s.projectStore.CreateProject(brigadeProject); err != nil {
		log.Println(
			errors.Wrap(err, "error creating new project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if err := s.oldProjectStore.CreateProject(brigadeProject); err != nil {
		log.Println(
			errors.Wrap(err, "error creating new project in old store"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	project = brignext.BrigadeProjectToBrigNextProject(brigadeProject)

	responseBytes, err := json.Marshal(project)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create project response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
