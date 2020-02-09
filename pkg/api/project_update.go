package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func (s *server) projectUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	if _, ok, err :=
		s.projectStore.GetProjectByName(name); err != nil {
		log.Println(
			errors.Wrapf(err, "error checking for existing project named %q", name),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
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
			errors.Wrap(err, "error unmarshaling body of update project request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if name != project.Name {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err := s.projectStore.UpdateProject(project); err != nil {
		log.Println(
			errors.Wrapf(err, "error updating project %q", project.Name),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
