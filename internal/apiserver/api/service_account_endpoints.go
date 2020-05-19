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

	responseBytes, err := json.Marshal(token)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseBytes)
}

func (s *server) serviceAccountList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	serviceAccountList, err := s.service.GetServiceAccounts(r.Context())
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all service accounts"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(serviceAccountList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list service accounts response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) serviceAccountGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	serviceAccount, err := s.service.GetServiceAccount(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrServiceAccountNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(serviceAccount)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) serviceAccountLock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if err := s.service.LockServiceAccount(r.Context(), id); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrServiceAccountNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error locking service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}

func (s *server) serviceAccountUnlock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	token, err := s.service.UnlockServiceAccount(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrServiceAccountNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error unlocking service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(token)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
