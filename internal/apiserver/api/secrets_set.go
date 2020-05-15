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
