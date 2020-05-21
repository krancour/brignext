package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *server) userList(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Users().List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (s *server) userGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Users().Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) userLock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Users().Lock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) userUnlock(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Users().Unlock(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}
