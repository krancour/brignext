package api

import (
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/http/filters/auth"
	"github.com/pkg/errors"
)

func (s *server) tokenDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	token := auth.UserTokenFromContext(r.Context())
	if token == "" {
		log.Println(
			"error: delete token request authenticated, but no token found in " +
				"request context",
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if err := s.userStore.DeleteUserToken(token); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting user token"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
