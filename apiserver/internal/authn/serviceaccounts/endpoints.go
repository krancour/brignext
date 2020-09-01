package serviceaccounts

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	serviceAccountSchemaLoader gojsonschema.JSONLoader
	service                    Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:              baseEndpoints,
		serviceAccountSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/service-account.json"),
		service:                    service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create service account
	router.HandleFunc(
		"/v2/service-accounts",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List service accounts
	router.HandleFunc(
		"/v2/service-accounts",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get service account
	router.HandleFunc(
		"/v2/service-accounts/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Lock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		e.TokenAuthFilter.Decorate(e.lock),
	).Methods(http.MethodPut)

	// Unlock service account
	router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		e.TokenAuthFilter.Decorate(e.unlock),
	).Methods(http.MethodDelete)

	// Grant role
	router.HandleFunc(
		"/v2/service-accounts/{id}/roles",
		e.TokenAuthFilter.Decorate(e.grantRole),
	).Methods(http.MethodPost)

	// Revoke role
	router.HandleFunc(
		"/v2/service-accounts/{id}/roles",
		e.TokenAuthFilter.Decorate(e.revokeRole),
	).Methods(http.MethodDelete)
}

func (e *endpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := authn.ServiceAccount{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.serviceAccountSchemaLoader,
			ReqBodyObj:          &serviceAccount,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Create(r.Context(), serviceAccount)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			e.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&core.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.List(
					r.Context(),
					authn.ServiceAccountsSelector{},
					opts,
				)
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
		})
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

func (e *endpoints) unlock(
	w http.ResponseWriter,
	r *http.Request,
) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) grantRole(w http.ResponseWriter, r *http.Request) {
	role := authn.Role{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &role,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.GrantRole(r.Context(), mux.Vars(r)["id"], role)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) revokeRole(w http.ResponseWriter, r *http.Request) {
	role := authn.Role{
		Name:  r.URL.Query().Get("name"),
		Scope: r.URL.Query().Get("scope"),
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.RevokeRole(r.Context(), mux.Vars(r)["id"], role)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
