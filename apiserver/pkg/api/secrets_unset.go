package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) secretsUnset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]
	workerName := mux.Vars(r)["workerName"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of unset secrets request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	keysStruct := struct {
		Keys []string `json:"keys"`
	}{}
	if err := json.Unmarshal(bodyBytes, &keysStruct); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of unset secrets request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err := s.service.UnsetSecrets(
		r.Context(),
		projectID,
		workerName,
		keysStruct.Keys,
	); err != nil {
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
