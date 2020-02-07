package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
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

	serviceAccount := &brignext.ServiceAccount{}
	if err := json.Unmarshal(bodyBytes, serviceAccount); err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error unmarshaling body of create service account request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// TODO: We should do some kind of validation!

	if existingServiceAccount, err :=
		s.userStore.GetServiceAccountByName(serviceAccount.Name); err != nil {
		log.Println(
			errors.Wrap(err, "error checking for existing service account"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if existingServiceAccount != nil {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	_, token, err := s.userStore.CreateServiceAccount(
		serviceAccount.Name,
		serviceAccount.Description,
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error creating new service account"),
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

	s.writeResponse(w, http.StatusOK, responseBytes)
}
