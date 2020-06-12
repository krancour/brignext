package serviceaccounts

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/pkg/api"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*api.BaseEndpoints
	serviceAccountSchemaLoader gojsonschema.JSONLoader
	service                    Service
}

func NewEndpoints(
	baseEndpoints *api.BaseEndpoints,
	service Service,
) api.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:              baseEndpoints,
		serviceAccountSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/service-account.json"),
		service:                    service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create service account
	router.HandleFunc(
		"/v2/service-accounts",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List service accounts
	router.HandleFunc(
		"/v2/service-accounts",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get service account
	router.HandleFunc(
		"/v2/service-accounts/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Lock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		e.TokenAuthFilter.Decorate(e.lock),
	).Methods(http.MethodPut)

	// Unlock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		e.TokenAuthFilter.Decorate(e.unlock),
	).Methods(http.MethodDelete)
}

func (e *endpoints) CheckHealth(ctx context.Context) error {
	return e.service.CheckHealth(ctx)
}

func (e *endpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := brignext.ServiceAccount{}
	e.ServeRequest(api.Request{
		W:                   w,
		R:                   r,
		ReqBodySchemaLoader: e.serviceAccountSchemaLoader,
		ReqBodyObj:          &serviceAccount,
		EndpointLogic: func() (interface{}, error) {
			return e.service.Create(r.Context(), serviceAccount)
		},
		SuccessCode: http.StatusCreated,
	})
}

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(api.Request{
		W: w,
		R: r,
		EndpointLogic: func() (interface{}, error) {
			return e.service.List(r.Context())
		},
		SuccessCode: http.StatusOK,
	})
}

func (e *endpoints) get(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(api.Request{
		W: w,
		R: r,
		EndpointLogic: func() (interface{}, error) {
			return e.service.Get(r.Context(), mux.Vars(r)["id"])
		},
		SuccessCode: http.StatusOK,
	})
}

func (e *endpoints) lock(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(api.Request{
		W: w,
		R: r,
		EndpointLogic: func() (interface{}, error) {
			return nil, e.service.Lock(r.Context(), mux.Vars(r)["id"])
		},
		SuccessCode: http.StatusOK,
	})
}

func (e *endpoints) unlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	e.ServeRequest(api.Request{
		W: w,
		R: r,
		EndpointLogic: func() (interface{}, error) {
			return e.service.Unlock(r.Context(), mux.Vars(r)["id"])
		},
		SuccessCode: http.StatusOK,
	})
}