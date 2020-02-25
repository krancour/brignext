package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

func (s *server) eventUpdateWorkerStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]

	workerName := mux.Vars(r)["workerName"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of update event request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	status := struct {
		Status brignext.WorkerStatus `json:"status"`
	}{}
	if err := json.Unmarshal(bodyBytes, &status); err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error unmarshaling body of update event worker status request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err :=
		s.service.UpdateEventWorkerStatus(
			r.Context(),
			eventID,
			workerName,
			status.Status,
		); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrWorkerNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error updating status on worker %q of event %q",
				workerName,
				eventID,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)

}
