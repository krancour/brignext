package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
)

func (s *server) workerLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]

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
		var logEntriesList brignext.LogEntryList
		var err error
		if init {
			logEntriesList, err = s.service.GetWorkerInitLogs(
				r.Context(),
				eventID,
			)
		} else {
			logEntriesList, err = s.service.GetWorkerLogs(
				r.Context(),
				eventID,
			)
		}
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(
					err,
					"error retrieving event %q worker logs",
					eventID,
				),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(logEntriesList)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling get worker logs response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	var logEntryCh <-chan brignext.LogEntry
	var err error
	if init {
		logEntryCh, err = s.service.StreamWorkerInitLogs(
			r.Context(),
			eventID,
		)
	} else {
		logEntryCh, err = s.service.StreamWorkerLogs(
			r.Context(),
			eventID,
		)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q worker",
				eventID,
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
