package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/xeipuuv/gojsonschema"
)

type serviceAccountEndpoints struct {
	*baseEndpoints
	serviceAccountSchemaLoader gojsonschema.JSONLoader
	service                    service.ServiceAccountsService
}

func (s *serviceAccountEndpoints) register(router *mux.Router) {
	// Create service account
	router.HandleFunc(
		"/v2/service-accounts",
		s.tokenAuthFilter.Decorate(s.create),
	).Methods(http.MethodPost)

	// List service accounts
	router.HandleFunc(
		"/v2/service-accounts",
		s.tokenAuthFilter.Decorate(s.list),
	).Methods(http.MethodGet)

	// Get service account
	router.HandleFunc(
		"/v2/service-accounts/{id}",
		s.tokenAuthFilter.Decorate(s.get),
	).Methods(http.MethodGet)

	// Lock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		s.tokenAuthFilter.Decorate(s.lock),
	).Methods(http.MethodPut)

	// Unlock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		s.tokenAuthFilter.Decorate(s.unlock),
	).Methods(http.MethodDelete)
}

func (s *serviceAccountEndpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := brignext.ServiceAccount{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.serviceAccountSchemaLoader,
		reqBodyObj:          &serviceAccount,
		endpointLogic: func() (interface{}, error) {
			return s.service.Create(r.Context(), serviceAccount)
		},
		successCode: http.StatusCreated,
	})
}

func (s *serviceAccountEndpoints) list(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (s *serviceAccountEndpoints) get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *serviceAccountEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Lock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *serviceAccountEndpoints) unlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Unlock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}
