package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountLock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if ok, err := s.service.LockServiceAccount(r.Context(), id); err != nil {
		log.Println(
			errors.Wrapf(err, "error locking service account %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseEmptyJSON)
}
