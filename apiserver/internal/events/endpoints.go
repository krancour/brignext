package events

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	eventSchemaLoader        gojsonschema.JSONLoader
	workerStatusSchemaLoader gojsonschema.JSONLoader
	jobSpecSchemaLoader      gojsonschema.JSONLoader
	jobStatusSchemaLoader    gojsonschema.JSONLoader
	service                  Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:            baseEndpoints,
		eventSchemaLoader:        gojsonschema.NewReferenceLoader("file:///brignext/schemas/event.json"),
		workerStatusSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/worker-status.json"),
		jobSpecSchemaLoader:      gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-spec.json"),
		jobStatusSchemaLoader:    gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-status.json"),
		service:                  service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create event
	router.HandleFunc(
		"/v2/events",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List events
	router.HandleFunc(
		"/v2/events",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get event
	router.HandleFunc(
		"/v2/events/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Cancel event
	router.HandleFunc(
		"/v2/events/{id}/cancellation",
		e.TokenAuthFilter.Decorate(e.cancel),
	).Methods(http.MethodPut)

	// Cancel a collection of events
	router.HandleFunc(
		"/v2/events/cancellations",
		e.TokenAuthFilter.Decorate(e.cancelMany),
	).Methods(http.MethodPost)

	// Delete event
	router.HandleFunc(
		"/v2/events/{id}",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// Delete a collection of events
	router.HandleFunc(
		"/v2/events",
		e.TokenAuthFilter.Decorate(e.deleteMany),
	).Methods(http.MethodDelete)

	// Start worker
	router.HandleFunc(
		"/v2/events/{id}/worker/start",
		e.TokenAuthFilter.Decorate(e.startWorker),
	).Methods(http.MethodPut)

	// Get/stream worker status
	router.HandleFunc(
		"/v2/events/{id}/worker/status",
		e.TokenAuthFilter.Decorate(e.getOrStreamWorkerStatus),
	).Methods(http.MethodGet)

	// Update worker status
	router.HandleFunc(
		"/v2/events/{id}/worker/status",
		e.TokenAuthFilter.Decorate(e.updateWorkerStatus),
	).Methods(http.MethodPut)

	// Create and start job
	router.HandleFunc(
		"/v2/events/{id}/worker/jobs/{jobName}/spec",
		e.TokenAuthFilter.Decorate(e.createJob),
	).Methods(http.MethodPut)

	// Get/stream job status
	router.HandleFunc(
		"/v2/events/{id}/worker/jobs/{jobName}/status",
		e.TokenAuthFilter.Decorate(e.getOrStreamJobStatus),
	).Methods(http.MethodGet)

	// Update job status
	router.HandleFunc(
		"/v2/events/{id}/worker/jobs/{jobName}/status",
		e.TokenAuthFilter.Decorate(e.updateJobStatus),
	).Methods(http.MethodPut)

	// Get/stream logs
	router.HandleFunc(
		"/v2/events/{id}/logs",
		e.TokenAuthFilter.Decorate(e.getOrStreamLogs),
	).Methods(http.MethodGet)
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	event := brignext.Event{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.eventSchemaLoader,
			ReqBodyObj:          &event,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Create(r.Context(), event)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	selector := brignext.EventsSelector{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			e.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&brignext.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}

	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]brignext.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = brignext.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context(), selector, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) get(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) cancel(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Cancel(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) cancelMany(
	w http.ResponseWriter,
	r *http.Request,
) {
	selector := brignext.EventsSelector{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]brignext.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = brignext.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.CancelMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) deleteMany(w http.ResponseWriter, r *http.Request) {
	selector := brignext.EventsSelector{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]brignext.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = brignext.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.DeleteMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) startWorker(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.StartWorker(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) getOrStreamWorkerStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	if !watch {
		e.ServeRequest(
			apimachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return e.service.GetWorkerStatus(r.Context(), id)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := e.service.WatchWorkerStatus(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			e.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving worker status stream for event %q", id),
		)
		e.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&brignext.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for status := range statusCh {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling worker status"))
			return
		}
		fmt.Fprint(w, string(statusBytes))
		w.(http.Flusher).Flush()
	}
}

func (e *endpoints) updateWorkerStatus(w http.ResponseWriter, r *http.Request) {
	status := brignext.WorkerStatus{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.workerStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil,
					e.service.UpdateWorkerStatus(r.Context(), mux.Vars(r)["id"], status)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) createJob(w http.ResponseWriter, r *http.Request) {
	jobSpec := brignext.JobSpec{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.jobSpecSchemaLoader,
			ReqBodyObj:          &jobSpec,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.CreateJob(
					r.Context(),
					mux.Vars(r)["id"],
					mux.Vars(r)["jobName"],
					jobSpec,
				)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) getOrStreamJobStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	if !watch {
		e.ServeRequest(
			apimachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return e.service.GetJobStatus(r.Context(), id, jobName)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := e.service.WatchJobStatus(r.Context(), id, jobName)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			e.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving job status stream for event %q job %q", id, jobName),
		)
		e.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&brignext.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for status := range statusCh {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling job status"))
			return
		}
		fmt.Fprint(w, string(statusBytes))
		w.(http.Flusher).Flush()
	}
}

func (e *endpoints) updateJobStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := brignext.JobStatus{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.jobStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.UpdateJobStatus(
					r.Context(),
					mux.Vars(r)["id"],
					mux.Vars(r)["jobName"],
					status,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) getOrStreamLogs(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	// nolint: errcheck
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))

	selector := brignext.LogsSelector{
		Job:       r.URL.Query().Get("job"),
		Container: r.URL.Query().Get("container"),
	}
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			e.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&brignext.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}

	if !stream {
		e.ServeRequest(
			apimachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return e.service.GetLogs(r.Context(), id, selector, opts)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	logEntryCh, err := e.service.StreamLogs(r.Context(), id, selector)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrNotFound); ok {
			e.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving log stream for event %q", id),
		)
		e.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&brignext.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}
