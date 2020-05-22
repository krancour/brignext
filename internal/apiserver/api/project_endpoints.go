package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
)

func (s *server) projectCreate(w http.ResponseWriter, r *http.Request) {
	project := brignext.Project{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.projectSchemaLoader,
		reqBodyObj:          &project,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Projects().Create(r.Context(), project)
		},
		successCode: http.StatusCreated,
	})
}

func (s *server) projectList(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Projects().List(r.Context())
		},
		successCode: http.StatusOK,
	})
}

func (s *server) projectGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Projects().Get(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) projectUpdate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	project := brignext.Project{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.projectSchemaLoader,
		reqBodyObj:          &project,
		endpointLogic: func() (interface{}, error) {
			if id != project.ID {
				return nil, brignext.NewErrBadRequest(
					"The project IDs in the URL path and request body do not match.",
				)
			}
			return nil, s.service.Projects().Update(r.Context(), project)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) projectDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Projects().Delete(r.Context(), id)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) secretsList(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return s.service.Projects().ListSecrets(r.Context(), projectID)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) secretSet(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	secret := brignext.Secret{}
	s.serveAPIRequest(apiRequest{
		w:                   w,
		r:                   r,
		reqBodySchemaLoader: s.secretSchemaLoader,
		reqBodyObj:          &secret,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Projects().SetSecret(
				r.Context(),
				projectID,
				secret,
			)
		},
		successCode: http.StatusOK,
	})
}

func (s *server) secretUnset(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["projectID"]
	secretID := mux.Vars(r)["secretID"]
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Projects().UnsetSecret(
				r.Context(),
				projectID,
				secretID,
			)
		},
		successCode: http.StatusOK,
	})
}
