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

func (s *server) eventCreate(w http.ResponseWriter, r *http.Request) {
	event := brignext.Event{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.eventSchemaLoader,
		reqBodyObj:          &event,
		endpointLogic: func() (interface{}, error) {
			return s.service.Events().Create(r.Context(), event)
		},
		successCode: http.StatusCreated,
	})
}

func (s *server) eventList(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			if projectID := r.URL.Query().Get("projectID"); projectID != "" {
				return s.service.Events().ListByProject(r.Context(), projectID)
			}
			return s.service.Events().List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (s *server) eventGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Events().Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) eventsCancel(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]
	// nolint: errcheck
	cancelRunning, _ := strconv.ParseBool(r.URL.Query().Get("cancelRunning"))
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			if eventID != "" {
				return s.service.Events().Cancel(r.Context(), eventID, cancelRunning)
			}
			return s.service.Events().CancelByProject(
				r.Context(),
				projectID,
				cancelRunning,
			)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) eventsDelete(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]
	// nolint: errcheck
	deletePending, _ := strconv.ParseBool(r.URL.Query().Get("deletePending"))
	// nolint: errcheck
	deleteRunning, _ := strconv.ParseBool(r.URL.Query().Get("deleteRunning"))
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			if eventID != "" {
				return s.service.Events().Delete(
					r.Context(),
					eventID,
					deletePending,
					deleteRunning,
				)
			}
			return s.service.Events().DeleteByProject(
				r.Context(),
				projectID,
				deletePending,
				deleteRunning,
			)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) workerUpdateStatus(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	status := brignext.WorkerStatus{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.workerStatusSchemaLoader,
		reqBodyObj:          &status,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Events().UpdateWorkerStatus(
				r.Context(),
				eventID,
				status,
			)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) workerLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	// nolint: errchecks
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))
	// nolint: errcheck
	init, _ := strconv.ParseBool(r.URL.Query().Get("init"))

	if !stream {
		s.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				if init {
					return s.service.Events().GetWorkerInitLogs(r.Context(), eventID)
				}
				return s.service.Events().GetWorkerLogs(r.Context(), eventID)
			},
			successCode: http.StatusOK,
		})
		return
	}

	var logEntryCh <-chan brignext.LogEntry
	var err error
	if init {
		logEntryCh, err = s.service.Events().StreamWorkerInitLogs(
			r.Context(),
			eventID,
		)
	} else {
		logEntryCh, err = s.service.Events().StreamWorkerLogs(
			r.Context(),
			eventID,
		)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			s.writeAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q worker",
				eventID,
			),
		)
		s.writeAPIResponse(
			w,
			http.StatusInternalServerError,
			brignext.NewErrInternalServer(),
		)
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

func (s *server) jobUpdateStatus(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	status := brignext.JobStatus{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.jobStatusSchemaLoader,
		reqBodyObj:          &status,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Events().UpdateJobStatus(
				r.Context(),
				eventID,
				jobName,
				status,
			)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) jobLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))
	// nolint: errcheck
	init, _ := strconv.ParseBool(r.URL.Query().Get("init"))

	if !stream {
		s.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				if init {
					return s.service.Events().GetJobInitLogs(
						r.Context(),
						eventID,
						jobName,
					)
				}
				return s.service.Events().GetJobLogs(
					r.Context(),
					eventID,
					jobName,
				)
			},
			successCode: http.StatusOK,
		})
		return
	}

	logEntryCh, err := s.service.Events().StreamJobLogs(
		r.Context(),
		eventID,
		jobName,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			s.writeAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q job %q",
				eventID,
				jobName,
			),
		)
		s.writeAPIResponse(
			w,
			http.StatusInternalServerError,
			brignext.NewErrInternalServer(),
		)
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
