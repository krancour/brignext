package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type projectsEndpoints struct {
	*restmachinery.BaseEndpoints
	projectSchemaLoader gojsonschema.JSONLoader
	service             core.ProjectsService
}

func NewProjectsEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service core.ProjectsService,
) restmachinery.Endpoints {
	// nolint: lll
	return &projectsEndpoints{
		BaseEndpoints:       baseEndpoints,
		projectSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		service:             service,
	}
}

func (p *projectsEndpoints) Register(router *mux.Router) {
	// Create Project
	router.HandleFunc(
		"/v2/projects",
		p.TokenAuthFilter.Decorate(p.create),
	).Methods(http.MethodPost)

	// List Projects
	router.HandleFunc(
		"/v2/projects",
		p.TokenAuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Get Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.get),
	).Methods(http.MethodGet)

	// Update Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.update),
	).Methods(http.MethodPut)

	// Delete Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.delete),
	).Methods(http.MethodDelete)
}

func (p *projectsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				return p.service.Create(r.Context(), project)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *projectsEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			p.WriteAPIResponse(
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
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.service.List(r.Context(), core.ProjectsSelector{}, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) get(w http.ResponseWriter, r *http.Request) {
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) update(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				if mux.Vars(r)["id"] != project.ID {
					return nil, &meta.ErrBadRequest{
						Reason: "The project IDs in the URL path and request body do " +
							"not match.",
					}
				}
				return p.service.Update(r.Context(), project)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
