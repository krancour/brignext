package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) userCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(errors.Wrap(err, "error reading body of create user request"))
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create user request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// TODO: We should do some kind of validation!

	if existingUser, err := s.userStore.GetUser(req.Username); err != nil {
		log.Println(
			errors.Wrap(err, "error checking for existing user"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if existingUser != nil {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err := s.userStore.CreateUser(req.Username, req.Password); err != nil {
		log.Println(
			errors.Wrap(err, "error creating new user"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
