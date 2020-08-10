package projects

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	projectSchemaLoader gojsonschema.JSONLoader
	secretSchemaLoader  gojsonschema.JSONLoader
	service             Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:       baseEndpoints,
		projectSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		secretSchemaLoader:  gojsonschema.NewReferenceLoader("file:///brignext/schemas/secret.json"),
		service:             service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create project
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List projects
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Update project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.update),
	).Methods(http.MethodPut)

	// Delete project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// List secrets
	router.HandleFunc(
		"/v2/projects/{id}/secrets",
		e.TokenAuthFilter.Decorate(e.listSecrets),
	).Methods(http.MethodGet)

	// Set secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.setSecret),
	).Methods(http.MethodPut)

	// Unset secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.unsetSecret),
	).Methods(http.MethodDelete)
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	project := brignext.Project{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Create(r.Context(), project)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context(), brignext.ProjectListOptions{})
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

func (e *endpoints) update(w http.ResponseWriter, r *http.Request) {
	project := brignext.Project{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				if mux.Vars(r)["id"] != project.ID {
					return nil, &brignext.ErrBadRequest{
						Reason: "The project IDs in the URL path and request body do " +
							"not match.",
					}
				}
				return e.service.Update(r.Context(), project)
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

func (e *endpoints) listSecrets(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.ListSecrets(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) setSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	secret := brignext.Secret{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.secretSchemaLoader,
			ReqBodyObj:          &secret,
			EndpointLogic: func() (interface{}, error) {
				if key != secret.Key {
					return nil, &brignext.ErrBadRequest{
						Reason: "The secret key in the URL path and request body do not " +
							"match.",
					}
				}
				return nil, e.service.SetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					secret,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) unsetSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.UnsetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					key,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
