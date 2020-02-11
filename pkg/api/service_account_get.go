package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	serviceAccount, ok, err := s.userStore.GetServiceAccount(
		storage.GetServiceAccountCriteria{
			ServiceAccountID: id,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}
	if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
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
