package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/xeipuuv/gojsonschema"
)

type projectEndpoints struct {
	*baseEndpoints
	projectSchemaLoader gojsonschema.JSONLoader
	secretSchemaLoader  gojsonschema.JSONLoader
	service             service.ProjectsService
}

func (p *projectEndpoints) register(router *mux.Router) {
	// Create project
	router.HandleFunc(
		"/v2/projects",
		p.tokenAuthFilter.Decorate(p.create),
	).Methods(http.MethodPost)

	// List projects
	router.HandleFunc(
		"/v2/projects",
		p.tokenAuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Get project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.tokenAuthFilter.Decorate(p.get),
	).Methods(http.MethodGet)

	// Update project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.tokenAuthFilter.Decorate(p.update),
	).Methods(http.MethodPut)

	// Delete project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.tokenAuthFilter.Decorate(p.delete),
	).Methods(http.MethodDelete)

	// List secrets
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets",
		p.tokenAuthFilter.Decorate(p.listSecrets),
	).Methods(http.MethodGet)

	// Set secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{secretID}",
		p.tokenAuthFilter.Decorate(p.setSecret),
	).Methods(http.MethodPut)

	// Unset secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{secretID}",
		p.tokenAuthFilter.Decorate(p.unsetSecret),
	).Methods(http.MethodDelete)
}

func (p *projectEndpoints) create(w http.ResponseWriter, r *http.Request) {
	project := brignext.Project{}
	p.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: p.projectSchemaLoader,
		reqBodyObj:          &project,
		endpointLogic: func() (interface{}, error) {
			return nil, p.service.Create(r.Context(), project)
		},
		successCode: http.StatusCreated,
	})
}

func (p *projectEndpoints) list(w http.ResponseWriter, r *http.Request) {
	p.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return p.service.List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return p.service.Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	project := brignext.Project{}
	p.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: p.projectSchemaLoader,
		reqBodyObj:          &project,
		endpointLogic: func() (interface{}, error) {
			if id != project.ID {
				return nil, brignext.NewErrBadRequest(
					"The project IDs in the URL path and request body do not match.",
				)
			}
			return nil, p.service.Update(r.Context(), project)
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	p.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, p.service.Delete(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) listSecrets(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	p.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return p.service.ListSecrets(r.Context(), projectID)
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) setSecret(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	secret := brignext.Secret{}
	p.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: p.secretSchemaLoader,
		reqBodyObj:          &secret,
		endpointLogic: func() (interface{}, error) {
			return nil, p.service.SetSecret(
				r.Context(),
				projectID,
				secret,
			)
		},
		successCode: http.StatusOK,
	})
}

func (p *projectEndpoints) unsetSecret(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	secretID := mux.Vars(r)["secretID"]
	p.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, p.service.UnsetSecret(
				r.Context(),
				projectID,
				secretID,
			)
		},
		successCode: http.StatusOK,
	})
}
