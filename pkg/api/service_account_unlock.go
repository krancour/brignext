package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountUnlock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	token, ok, err := s.service.UnlockServiceAccount(r.Context(), id)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error unlocking service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Token string `json:"token"`
		}{
			Token: token,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create service account response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
