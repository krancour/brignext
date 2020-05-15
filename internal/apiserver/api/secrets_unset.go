package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) secretUnset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectID := mux.Vars(r)["projectID"]
	secretID := mux.Vars(r)["secretID"]

	if err :=
		s.service.UnsetSecret(r.Context(), projectID, secretID); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error unsetting secrets"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
	return
}
