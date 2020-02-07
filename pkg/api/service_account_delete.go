package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) serviceAccountDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	name := mux.Vars(r)["name"]

	if err := s.userStore.DeleteServiceAccountByName(name); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting service account"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
