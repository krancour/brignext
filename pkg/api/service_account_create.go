package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
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

	serviceAccount := brignext.ServiceAccount{}
	if err := json.Unmarshal(bodyBytes, &serviceAccount); err != nil {
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

	if _, ok, err :=
		s.userStore.GetServiceAccount(serviceAccount.Name); err != nil {
		log.Println(
			errors.Wrapf(
				err,
				"error checking for existing service account named %q",
				serviceAccount.Name,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if ok {
		s.writeResponse(w, http.StatusConflict, responseEmptyJSON)
		return
	}

	serviceAccount.Token = crypto.NewToken(256)
	serviceAccount.Created = time.Now()

	if err := s.userStore.CreateServiceAccount(serviceAccount); err != nil {
		log.Println(
			errors.Wrapf(
				err,
				"error creating new service account %q",
				serviceAccount.Name,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Token string `json:"token"`
		}{
			Token: serviceAccount.Token,
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
