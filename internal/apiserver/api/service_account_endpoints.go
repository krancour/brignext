package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
)

func (s *server) serviceAccountCreate(w http.ResponseWriter, r *http.Request) {
	serviceAccount := brignext.ServiceAccount{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.serviceAccountSchemaLoader,
		reqBodyObj:          &serviceAccount,
		endpointLogic: func() (interface{}, error) {
			return s.service.ServiceAccounts().Create(r.Context(), serviceAccount)
		},
		successCode: http.StatusCreated,
	})
}

func (s *server) serviceAccountList(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.ServiceAccounts().List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (s *server) serviceAccountGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.ServiceAccounts().Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) serviceAccountLock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.ServiceAccounts().Lock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) serviceAccountUnlock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.ServiceAccounts().Unlock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}
