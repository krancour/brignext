package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type JobsEndpoints struct {
	*restmachinery.BaseEndpoints
	JobSchemaLoader       gojsonschema.JSONLoader
	JobStatusSchemaLoader gojsonschema.JSONLoader
	Service               core.JobsService
}

func (j *JobsEndpoints) Register(router *mux.Router) {
	// Create job
	router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}",
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

func (j *JobsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	job := core.Job{}
	j.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.JobSchemaLoader,
			ReqBodyObj:          &job,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Create(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
					job,
				)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (j *JobsEndpoints) start(w http.ResponseWriter, r *http.Request) {
	j.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.Start(
					r.Context(),
					mux.Vars(r)["eventID"],
					mux.Vars(r)["jobName"],
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (j *JobsEndpoints) getOrStreamStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]
	// nolint: errcheck
	watch, _ := strconv.ParseBool(r.URL.Query().Get("watch"))

	if !watch {
		j.ServeRequest(
			restmachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					return j.Service.GetStatus(r.Context(), id, jobName)
				},
				SuccessCode: http.StatusOK,
			},
		)
		return
	}

	statusCh, err := j.Service.WatchStatus(r.Context(), id, jobName)
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

func (j *JobsEndpoints) updateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	status := core.JobStatus{}
	j.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: j.JobStatusSchemaLoader,
			ReqBodyObj:          &status,
			EndpointLogic: func() (interface{}, error) {
				return nil, j.Service.UpdateStatus(
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
