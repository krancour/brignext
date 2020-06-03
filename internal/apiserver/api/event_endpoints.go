package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type eventEndpoints struct {
	*baseEndpoints
	eventSchemaLoader        gojsonschema.JSONLoader
	jobStatusSchemaLoader    gojsonschema.JSONLoader
	workerStatusSchemaLoader gojsonschema.JSONLoader
	service                  service.EventsService
}

func (e *eventEndpoints) register(router *mux.Router) {
	// Create event
	router.HandleFunc(
		"/v2/events",
		e.tokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List events
	router.HandleFunc(
		"/v2/events",
		e.tokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get event
	router.HandleFunc(
		"/v2/events/{id}",
		e.tokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Cancel event
	router.HandleFunc(
		"/v2/events/{id}/cancellation",
		e.tokenAuthFilter.Decorate(e.cancel),
	).Methods(http.MethodPut)

	// Cancel a collection of events
	router.HandleFunc(
		"/v2/events/cancellations",
		e.tokenAuthFilter.Decorate(e.cancelCollection),
	).Methods(http.MethodPost)

	// Delete event
	router.HandleFunc(
		"/v2/events/{id}",
		e.tokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// Delete a collection of events
	router.HandleFunc(
		"/v2/events",
		e.tokenAuthFilter.Decorate(e.deleteCollection),
	).Methods(http.MethodDelete)

	// Update worker status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/status",
		e.tokenAuthFilter.Decorate(e.updateWorkerStatus),
	).Methods(http.MethodPut)

	// Get/stream worker logs
	router.HandleFunc(
		"/v2/events/{eventID}/worker/logs",
		e.tokenAuthFilter.Decorate(e.getOrStreamWorkerLogs),
	).Methods(http.MethodGet)

	// Update job status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		e.tokenAuthFilter.Decorate(e.updateJobStatus),
	).Methods(http.MethodPut)

	// Get/stream job logs
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/logs",
		e.tokenAuthFilter.Decorate(e.getOrStreamJobLogs),
	).Methods(http.MethodGet)
}

func (e *eventEndpoints) create(w http.ResponseWriter, r *http.Request) {
	event := brignext.Event{}
	e.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: e.eventSchemaLoader,
		reqBodyObj:          &event,
		endpointLogic: func() (interface{}, error) {
			return e.service.Create(r.Context(), event)
		},
		successCode: http.StatusCreated,
	})
}

func (e *eventEndpoints) list(w http.ResponseWriter, r *http.Request) {
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			if projectID := r.URL.Query().Get("projectID"); projectID != "" {
				return e.service.ListByProject(r.Context(), projectID)
			}
			return e.service.List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return e.service.Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) cancel(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, e.service.Cancel(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) cancelCollection(
	w http.ResponseWriter,
	r *http.Request,
) {
	opts := brignext.EventListOptions{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		opts.WorkerPhases = make([]brignext.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			opts.WorkerPhases[i] = brignext.WorkerPhase(workerPhaseStr)
		}
	}
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return e.service.CancelCollection(r.Context(), opts)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, e.service.Delete(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) deleteCollection(
	w http.ResponseWriter,
	r *http.Request,
) {
	opts := brignext.EventListOptions{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		opts.WorkerPhases = make([]brignext.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			opts.WorkerPhases[i] = brignext.WorkerPhase(workerPhaseStr)
		}
	}
	e.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return e.service.DeleteCollection(r.Context(), opts)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) updateWorkerStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := mux.Vars(r)["eventID"]
	status := brignext.WorkerStatus{}
	e.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: e.workerStatusSchemaLoader,
		reqBodyObj:          &status,
		endpointLogic: func() (interface{}, error) {
			return nil, e.service.UpdateWorkerStatus(
				r.Context(),
				eventID,
				status,
			)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) getOrStreamWorkerLogs(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := mux.Vars(r)["eventID"]
	// nolint: errchecks
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))
	// nolint: errcheck
	init, _ := strconv.ParseBool(r.URL.Query().Get("init"))

	if !stream {
		e.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				if init {
					return e.service.GetWorkerInitLogs(r.Context(), eventID)
				}
				return e.service.GetWorkerLogs(r.Context(), eventID)
			},
			successCode: http.StatusOK,
		})
		return
	}

	var logEntryCh <-chan brignext.LogEntry
	var err error
	if init {
		logEntryCh, err = e.service.StreamWorkerInitLogs(
			r.Context(),
			eventID,
		)
	} else {
		logEntryCh, err = e.service.StreamWorkerLogs(
			r.Context(),
			eventID,
		)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			e.writeAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q worker",
				eventID,
			),
		)
		e.writeAPIResponse(
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

func (e *eventEndpoints) updateJobStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	status := brignext.JobStatus{}
	e.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: e.jobStatusSchemaLoader,
		reqBodyObj:          &status,
		endpointLogic: func() (interface{}, error) {
			return nil, e.service.UpdateJobStatus(
				r.Context(),
				eventID,
				jobName,
				status,
			)
		},
		successCode: http.StatusOK,
	})
}

func (e *eventEndpoints) getOrStreamJobLogs(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))
	// nolint: errcheck
	init, _ := strconv.ParseBool(r.URL.Query().Get("init"))

	if !stream {
		e.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				if init {
					return e.service.GetJobInitLogs(
						r.Context(),
						eventID,
						jobName,
					)
				}
				return e.service.GetJobLogs(
					r.Context(),
					eventID,
					jobName,
				)
			},
			successCode: http.StatusOK,
		})
		return
	}

	logEntryCh, err := e.service.StreamJobLogs(
		r.Context(),
		eventID,
		jobName,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			e.writeAPIResponse(w, http.StatusNotFound, errors.Cause(err))
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
		e.writeAPIResponse(
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
