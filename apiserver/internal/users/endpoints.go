package users

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
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
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			var err error
			if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
				opts.Limit < 1 || opts.Limit > 100 {
				e.WriteAPIResponse(
					w,
					http.StatusBadRequest,
					&brignext.ErrBadRequest{
						Reason: fmt.Sprintf(
							`Invalid value %q for "limit" query parameter`,
							limitStr,
						),
					},
				)
			}
			return
		}
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(r.Context(), brignext.UsersSelector{}, opts)
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
