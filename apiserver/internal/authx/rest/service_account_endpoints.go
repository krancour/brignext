package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type serviceAccountEndpoints struct {
	*restmachinery.BaseEndpoints
	serviceAccountSchemaLoader gojsonschema.JSONLoader
	service                    authx.ServiceAccountsService
}

func NewServiceAccountEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service authx.ServiceAccountsService,
) restmachinery.Endpoints {
	// nolint: lll
	return &serviceAccountEndpoints{
		BaseEndpoints:              baseEndpoints,
		serviceAccountSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/service-account.json"),
		service:                    service,
	}
}

func (s *serviceAccountEndpoints) Register(router *mux.Router) {
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

func (s *serviceAccountEndpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := authx.ServiceAccount{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: s.serviceAccountSchemaLoader,
			ReqBodyObj:          &serviceAccount,
			EndpointLogic: func() (interface{}, error) {
				return s.service.Create(r.Context(), serviceAccount)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (s *serviceAccountEndpoints) list(w http.ResponseWriter, r *http.Request) {
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
				return s.service.List(
					r.Context(),
					authx.ServiceAccountsSelector{},
					opts,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *serviceAccountEndpoints) get(w http.ResponseWriter, r *http.Request) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		})
}

func (s *serviceAccountEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.Lock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *serviceAccountEndpoints) unlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
