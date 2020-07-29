package users

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/api"
	"github.com/pkg/errors"
)

type endpoints struct {
	*api.BaseEndpoints
	service Service
}

func NewEndpoints(
	baseEndpoints *api.BaseEndpoints,
	service Service,
) api.Endpoints {
	return &endpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// List users
	router.HandleFunc(
		"/v2/users",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get user
	router.HandleFunc(
		"/v2/users/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Lock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		e.TokenAuthFilter.Decorate(e.lock),
	).Methods(http.MethodPut)

	// Unlock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		e.TokenAuthFilter.Decorate(e.unlock),
	).Methods(http.MethodDelete)
}

func (e *endpoints) CheckHealth(ctx context.Context) error {
	if err := e.service.CheckHealth(ctx); err != nil {
		return errors.Wrap(err, "error checking users service health")
	}
	return nil
}

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context())
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

func (e *endpoints) lock(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Lock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) unlock(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		api.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
