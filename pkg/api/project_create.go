package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
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

	if validationResult, err := gojsonschema.Validate(
		s.projectSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(errors.Wrap(err, "error validating request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	project := brignext.Project{}
	if err := json.Unmarshal(bodyBytes, &project); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if _, ok, err :=
		s.projectStore.GetProjectByName(project.Name); err != nil {
		log.Println(
			errors.Wrapf(
				err,
				"error checking for existing project named %q",
				project.Name,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if ok {
		s.writeResponse(w, http.StatusConflict, responseEmptyJSON)
		return
	}

	if _, err := s.projectStore.CreateProject(project); err != nil {
		log.Println(
			errors.Wrapf(err, "error creating project %q", project.Name),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseEmptyJSON)
}
