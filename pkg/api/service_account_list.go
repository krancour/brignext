package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

	responseServiceAccounts := make(
		[]struct {
			Name        string    `json:"name"`
			Description string    `json:"description"`
			Created     time.Time `json:"created"`
		},
		len(serviceAccounts),
	)
	for i, serviceAccount := range serviceAccounts {
		responseServiceAccounts[i].Name = serviceAccount.Name
		responseServiceAccounts[i].Description = serviceAccount.Description
		responseServiceAccounts[i].Created = serviceAccount.Created
	}

	responseBytes, err := json.Marshal(responseServiceAccounts)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list service accounts response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
