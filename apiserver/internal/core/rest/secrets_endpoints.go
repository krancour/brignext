package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type secretsEndpoints struct {
	*restmachinery.BaseEndpoints
	secretSchemaLoader gojsonschema.JSONLoader
	service            core.SecretsService
}

func NewSecretsEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service core.SecretsService,
) restmachinery.Endpoints {
	// nolint: lll
	return &secretsEndpoints{
		BaseEndpoints:      baseEndpoints,
		secretSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/secret.json"),
		service:            service,
	}
}

func (s *secretsEndpoints) Register(router *mux.Router) {
	// List Secrets
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets",
		s.TokenAuthFilter.Decorate(s.list),
	).Methods(http.MethodGet)

	// Set Secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{key}",
		s.TokenAuthFilter.Decorate(s.set),
	).Methods(http.MethodPut)

	// Unset Secret
	router.HandleFunc(
		"/v2/projects/{projectID}/secrets/{key}",
		s.TokenAuthFilter.Decorate(s.unset),
	).Methods(http.MethodDelete)
}

func (s *secretsEndpoints) list(w http.ResponseWriter, r *http.Request) {
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
				return s.service.List(r.Context(), mux.Vars(r)["projectID"], opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *secretsEndpoints) set(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	secret := core.Secret{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: s.secretSchemaLoader,
			ReqBodyObj:          &secret,
			EndpointLogic: func() (interface{}, error) {
				if key != secret.Key {
					return nil, &meta.ErrBadRequest{
						Reason: "The secret key in the URL path and request body do not " +
							"match.",
					}
				}
				return nil, s.service.Set(
					r.Context(),
					mux.Vars(r)["projectID"],
					secret,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *secretsEndpoints) unset(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.Unset(
					r.Context(),
					mux.Vars(r)["projectID"],
					key,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
