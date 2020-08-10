package users

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	service Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
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

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := brignext.UserListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
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

func (e *endpoints) lock(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
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
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
