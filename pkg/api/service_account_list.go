package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) serviceAccountList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	serviceAccounts, err := s.userStore.GetServiceAccounts()
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all service accounts"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(serviceAccounts)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list service accounts response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
