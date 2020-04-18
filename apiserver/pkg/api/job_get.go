package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/krancour/brignext"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) jobGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]
	workerName := mux.Vars(r)["workerName"]
	jobName := mux.Vars(r)["jobName"]

	job, err := s.service.GetJob(r.Context(), eventID, workerName, jobName)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		} else if _, ok := errors.Cause(err).(*brignext.ErrWorkerNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		} else if _, ok := errors.Cause(err).(*brignext.ErrJobNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving event %q worker %q job %q",
				eventID,
				workerName,
				jobName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(job)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get job response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}