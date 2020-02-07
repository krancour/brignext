package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func (s *server) userList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	users, err := s.userStore.GetUsers()
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving all users"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseUsers := make(
		[]struct {
			Username  string    `json:"username"`
			FirstSeen time.Time `json:"firstSeen"`
		},
		len(users),
	)
	for i, user := range users {
		responseUsers[i].Username = user.Username
		responseUsers[i].FirstSeen = user.FirstSeen
	}

	responseBytes, err := json.Marshal(responseUsers)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list users response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
