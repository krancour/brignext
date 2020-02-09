package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	serviceAccount, ok, err := s.userStore.GetServiceAccount(name)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving service account %q", name),
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
