package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type ServiceAccountEndpoints struct {
	*restmachinery.BaseEndpoints
	ServiceAccountSchemaLoader gojsonschema.JSONLoader
	Service                    authx.ServiceAccountsService
}

func (s *ServiceAccountEndpoints) Register(router *mux.Router) {
	// Create service account
	router.HandleFunc(
		"/v2/service-accounts",
		s.TokenAuthFilter.Decorate(s.create),
	).Methods(http.MethodPost)

	// List service accounts
	router.HandleFunc(
		"/v2/service-accounts",
		s.TokenAuthFilter.Decorate(s.list),
	).Methods(http.MethodGet)

	// Get service account
	router.HandleFunc(
		"/v2/service-accounts/{id}",
		s.TokenAuthFilter.Decorate(s.get),
	).Methods(http.MethodGet)

	// Lock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		s.TokenAuthFilter.Decorate(s.lock),
	).Methods(http.MethodPut)

	// Unlock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		s.TokenAuthFilter.Decorate(s.unlock),
	).Methods(http.MethodDelete)
}

func (s *ServiceAccountEndpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := authx.ServiceAccount{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: s.ServiceAccountSchemaLoader,
			ReqBodyObj:          &serviceAccount,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.Create(r.Context(), serviceAccount)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (s *ServiceAccountEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			s.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.List(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *ServiceAccountEndpoints) get(w http.ResponseWriter, r *http.Request) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		})
}

func (s *ServiceAccountEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.Service.Lock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *ServiceAccountEndpoints) unlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.Service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
