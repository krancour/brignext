package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/internal/api"
	errs "github.com/krancour/brignext/v2/internal/errors"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*api.BaseEndpoints
	eventSchemaLoader        gojsonschema.JSONLoader
	jobStatusSchemaLoader    gojsonschema.JSONLoader
	workerStatusSchemaLoader gojsonschema.JSONLoader
	service                  Service
}

func NewEndpoints(
	baseEndpoints *api.BaseEndpoints,
	service Service,
) api.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:            baseEndpoints,
		eventSchemaLoader:        gojsonschema.NewReferenceLoader("file:///brignext/schemas/event.json"),
		jobStatusSchemaLoader:    gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-status.json"),
		workerStatusSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/worker-status.json"),
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
		e.TokenAuthFilter.Decorate(e.cancelCollection),
	).Methods(http.MethodPost)

	// Delete event
	router.HandleFunc(
		"/v2/events/{id}",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// Delete a collection of events
	router.HandleFunc(
		"/v2/events",
		e.TokenAuthFilter.Decorate(e.deleteCollection),
	).Methods(http.MethodDelete)

	// Update worker status
	router.HandleFunc(
		"/v2/events/{id}/worker/status",
		e.TokenAuthFilter.Decorate(e.updateWorkerStatus),
	).Methods(http.MethodPut)

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

func (e *endpoints) CheckHealth(ctx context.Context) error {
	if err := e.service.CheckHealth(ctx); err != nil {
		return errors.Wrap(err, "error checking events service health")
	}
	return nil
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	event := brignext.Event{}
	e.ServeRequest(
		api.InboundRequest{
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
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) get(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
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
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Cancel(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) cancelCollection(
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
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.CancelCollection(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) deleteCollection(
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
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.DeleteCollection(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) updateWorkerStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := brignext.WorkerStatus{}
	e.ServeRequest(
		api.InboundRequest{
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

func (e *endpoints) getOrStreamLogs(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	// nolint: errcheck
	stream, _ := strconv.ParseBool(r.URL.Query().Get("stream"))

	opts := brignext.LogOptions{
		Job:       r.URL.Query().Get("job"),
		Container: r.URL.Query().Get("container"),
	}

	if !stream {
		e.ServeRequest(
			api.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return e.service.GetLogs(r.Context(), id, opts)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	logEntryCh, err := e.service.StreamLogs(r.Context(), id, opts)
	if err != nil {
		if _, ok := errors.Cause(err).(*errs.ErrNotFound); ok {
			e.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving log stream for event %q", id),
		)
		e.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			errs.NewErrInternalServer(),
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

func (e *endpoints) updateJobStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := brignext.JobStatus{}
	e.ServeRequest(
		api.InboundRequest{
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
