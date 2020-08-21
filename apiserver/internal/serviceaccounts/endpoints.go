package serviceaccounts

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
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
}

func (e *endpoints) create(
	w http.ResponseWriter,
	r *http.Request,
) {
	serviceAccount := brignext.ServiceAccount{}
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
				&brignext.ErrBadRequest{
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
					brignext.ServiceAccountsSelector{},
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
