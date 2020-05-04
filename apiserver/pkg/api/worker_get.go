package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) workerGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]
	workerName := mux.Vars(r)["workerName"]

	worker, err := s.service.GetWorker(r.Context(), eventID, workerName)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		} else if _, ok := errors.Cause(err).(*brignext.ErrWorkerNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving event %q worker %q",
				eventID,
				workerName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(worker)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get worker response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
