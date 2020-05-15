package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) userList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	userList, err := s.service.GetUsers(r.Context())
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all users"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(userList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list users response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
