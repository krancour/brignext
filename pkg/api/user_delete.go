package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) userDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	username := mux.Vars(r)["username"]

	if err := s.userStore.DeleteUser(username); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting user"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to sessions

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
