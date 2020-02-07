package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	serviceAccount, err := s.userStore.GetServiceAccountByName(name)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving service account"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if serviceAccount == nil {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Name        string    `json:"name"`
			Description string    `json:"description"`
			Created     time.Time `json:"created"`
		}{
			Name:        serviceAccount.Name,
			Description: serviceAccount.Description,
			Created:     serviceAccount.Created,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
