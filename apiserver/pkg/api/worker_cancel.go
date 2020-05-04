package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) workerCancel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]
	workerName := mux.Vars(r)["workerName"]

	cancelRunningStr := r.URL.Query().Get("cancelRunning")
	var cancelRunning bool
	if cancelRunningStr != "" {
		cancelRunning, _ = strconv.ParseBool(cancelRunningStr) // nolint: errcheck
	}

	canceled, err := s.service.CancelWorker(
		r.Context(),
		eventID,
		workerName,
		cancelRunning,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrWorkerNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error canceling event %q worker %q",
				eventID,
				workerName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			Canceled bool `json:"canceled"`
		}{
			Canceled: canceled,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling cancel worker response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
