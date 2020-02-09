package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if err := s.userStore.DeleteUser(id); err != nil {
		log.Println(
			errors.Wrapf(err, "error deleting user %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to sessions

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
