package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func (s *server) serviceAccountCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of create service account request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	var validationResult *gojsonschema.Result
	if validationResult, err = gojsonschema.Validate(
		s.serviceAccountSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(
			errors.Wrap(err, "error validating create service account request"),
		)
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

	serviceAccount := brignext.ServiceAccount{}
	if err = json.Unmarshal(bodyBytes, &serviceAccount); err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error unmarshaling body of create service account request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	token, err := s.service.CreateServiceAccount(r.Context(), serviceAccount)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrServiceAccountIDConflict); ok {
			s.writeResponse(w, http.StatusConflict, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error creating service account %q",
				serviceAccount.ID,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseBytes)
}
