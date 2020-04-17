package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

func (s *server) jobLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	workerName := mux.Vars(r)["workerName"]
	jobName := mux.Vars(r)["jobName"]

	streamStr := r.URL.Query().Get("stream")
	var stream bool
	if streamStr != "" {
		stream, _ = strconv.ParseBool(streamStr) // nolint: errcheck
	}

	initStr := r.URL.Query().Get("init")
	var init bool
	if initStr != "" {
		init, _ = strconv.ParseBool(initStr) // nolint: errcheck
	}

	if !stream {
		var logEntries []brignext.LogEntry
		var err error
		if init {
			logEntries, err = s.service.GetJobInitLogs(
				r.Context(),
				eventID,
				workerName,
				jobName,
			)
		} else {
			logEntries, err = s.service.GetJobLogs(
				r.Context(),
				eventID,
				workerName,
				jobName,
			)
		}
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
					"error retrieving event %q worker %q job %q logs",
					eventID,
					workerName,
				),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(logEntries)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling get job logs response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	logEntryCh, err := s.service.StreamJobLogs(
		r.Context(),
		eventID,
		workerName,
		jobName,
	)
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
				"error retrieving log stream for event %q worker %q job %q",
				eventID,
				workerName,
				jobName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error unmarshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}
