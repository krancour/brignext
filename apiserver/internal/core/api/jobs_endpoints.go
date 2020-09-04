package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type jobsEndpoints struct {
	*apimachinery.BaseEndpoints
	jobSpecSchemaLoader   gojsonschema.JSONLoader
	jobStatusSchemaLoader gojsonschema.JSONLoader
	service               core.EventsService
}

func NewJobsEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service core.EventsService,
) apimachinery.Endpoints {
	// nolint: lll
	return &jobsEndpoints{
		BaseEndpoints:         baseEndpoints,
		jobSpecSchemaLoader:   gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-spec.json"),
		jobStatusSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-status.json"),
		service:               service,
	}
}

func (j *jobsEndpoints) Register(router *mux.Router) {
	// Create job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/spec",
		j.TokenAuthFilter.Decorate(j.create),
	).Methods(http.MethodPut)

	// Start job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/start",
		j.TokenAuthFilter.Decorate(j.start),
	).Methods(http.MethodPut)

	// Get/stream job status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		j.TokenAuthFilter.Decorate(j.getOrStreamStatus),
	).Methods(http.MethodGet)

	// Update job status
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		j.TokenAuthFilter.Decorate(j.updateStatus),
	).Methods(http.MethodPut)
}

func (j *jobsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	jobSpec := core.JobSpec{}
	j.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.jobSpecSchemaLoader,
			ReqBodyObj:          &jobSpec,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.service.CreateJob(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
					jobSpec,
				)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (j *jobsEndpoints) start(w http.ResponseWriter, r *http.Request) {
	j.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.service.StartJob(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *jobsEndpoints) getOrStreamStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	if !watch {
		j.ServeRequest(
			apimachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return j.service.GetJobStatus(r.Context(), id, jobName)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := j.service.WatchJobStatus(r.Context(), id, jobName)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			j.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Printf(
			"error retrieving job status stream for event %q job %q: %s",
			id,
			jobName,
			err,
		)
		j.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&meta.ErrInternalServer{},
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

func (j *jobsEndpoints) updateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := core.JobStatus{}
	j.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.jobStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.service.UpdateJobStatus(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
					status,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
