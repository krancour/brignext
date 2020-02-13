package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userLock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if ok, err := s.service.LockUser(r.Context(), id); err != nil {
		log.Println(
			errors.Wrapf(err, "error locking user %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseEmptyJSON)
}
