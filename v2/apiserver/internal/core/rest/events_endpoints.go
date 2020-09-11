package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type eventsEndpoints struct {
	*restmachinery.BaseEndpoints
	eventSchemaLoader gojsonschema.JSONLoader
	service           core.EventsService
}

func NewEventsEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service core.EventsService,
) restmachinery.Endpoints {
	// nolint: lll
	return &eventsEndpoints{
		BaseEndpoints:     baseEndpoints,
		eventSchemaLoader: gojsonschema.NewReferenceLoader("file:///brigade/schemas/event.json"),
		service:           service,
	}
}

func (e *eventsEndpoints) Register(router *mux.Router) {
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
}

func (e *eventsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	event := core.Event{}
	e.ServeRequest(
		restmachinery.InboundRequest{
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

func (e *eventsEndpoints) list(w http.ResponseWriter, r *http.Request) {
	selector := core.EventsSelector{
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
				&meta.ErrBadRequest{
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
		selector.WorkerPhases = make([]core.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = core.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context(), selector, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsEndpoints) get(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsEndpoints) cancel(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Cancel(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsEndpoints) cancelMany(
	w http.ResponseWriter,
	r *http.Request,
) {
	selector := core.EventsSelector{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]core.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = core.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.CancelMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *eventsEndpoints) deleteMany(w http.ResponseWriter, r *http.Request) {
	selector := core.EventsSelector{
		ProjectID: r.URL.Query().Get("projectID"),
	}
	workerPhasesStr := r.URL.Query().Get("workerPhases")
	if workerPhasesStr != "" {
		workerPhaseStrs := strings.Split(workerPhasesStr, ",")
		selector.WorkerPhases = make([]core.WorkerPhase, len(workerPhaseStrs))
		for i, workerPhaseStr := range workerPhaseStrs {
			selector.WorkerPhases[i] = core.WorkerPhase(workerPhaseStr)
		}
	}
	e.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.DeleteMany(r.Context(), selector)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
