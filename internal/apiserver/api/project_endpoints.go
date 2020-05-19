package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
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
		log.Println(errors.Wrap(err, "error validating create project request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
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

	if err := s.service.CreateProject(r.Context(), project); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectIDConflict); ok {
			s.writeResponse(w, http.StatusConflict, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error creating project %q", project.ID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseEmptyJSON)
}

func (s *server) projectList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectList, err := s.service.GetProjects(r.Context())
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all projects"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(projectList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list projects response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) projectGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	project, err := s.service.GetProject(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving project %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(project)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get project response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) projectUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

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
		log.Println(errors.Wrap(err, "error validating update project request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
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

	if id != project.ID {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err := s.service.UpdateProject(r.Context(), project); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error updating project %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}

func (s *server) projectDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if err := s.service.DeleteProject(r.Context(), id); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error deleting project %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}

func (s *server) secretsList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]

	secretList, err := s.service.GetSecrets(r.Context(), projectID)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error retrieving secrets"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(secretList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list secrets response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
	return
}

func (s *server) secretSet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of set secret request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if validationResult, err := gojsonschema.Validate(
		s.secretSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(errors.Wrap(err, "error validating set secret request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	secret := brignext.Secret{}
	if err := json.Unmarshal(bodyBytes, &secret); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of set secret request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err := s.service.SetSecret(
		r.Context(),
		projectID,
		secret,
	); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error setting secret"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
	return
}

func (s *server) secretUnset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]
	secretID := mux.Vars(r)["secretID"]

	if err :=
		s.service.UnsetSecret(r.Context(), projectID, secretID); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error unsetting secrets"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
	return
}
