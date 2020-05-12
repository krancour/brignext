package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) userUnlock(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	if err := s.service.UnlockUser(r.Context(), id); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrUserNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error unlocking user %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
