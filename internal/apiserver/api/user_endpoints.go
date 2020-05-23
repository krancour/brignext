package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
)

type userEndpoints struct {
	*baseEndpoints
	service service.UsersService
}

func (u *userEndpoints) register(router *mux.Router) {
	// List users
	router.HandleFunc(
		"/v2/users",
		u.tokenAuthFilter.Decorate(u.list),
	).Methods(http.MethodGet)

	// Get user
	router.HandleFunc(
		"/v2/users/{id}",
		u.tokenAuthFilter.Decorate(u.get),
	).Methods(http.MethodGet)

	// Lock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.tokenAuthFilter.Decorate(u.lock),
	).Methods(http.MethodPut)

	// Unlock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.tokenAuthFilter.Decorate(u.unlock),
	).Methods(http.MethodDelete)
}

func (u *userEndpoints) list(w http.ResponseWriter, r *http.Request) {
	u.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return u.service.List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (u *userEndpoints) get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	u.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return u.service.Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (u *userEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	u.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, u.service.Lock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (u *userEndpoints) unlock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	u.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, u.service.Unlock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}
