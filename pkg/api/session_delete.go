package api

import (
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/http/filters/auth"
	"github.com/pkg/errors"
)

func (s *server) sessionDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	sessionID := auth.SessionIDFromContext(r.Context())
	if sessionID == "" {
		log.Println(
			"error: delete session request authenticated, but no session ID found " +
				"in request context",
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if err := s.service.DeleteSession(r.Context(), sessionID); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting session"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
