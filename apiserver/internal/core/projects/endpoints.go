package projects

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	projectSchemaLoader gojsonschema.JSONLoader
	secretSchemaLoader  gojsonschema.JSONLoader
	service             Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints:       baseEndpoints,
		projectSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		secretSchemaLoader:  gojsonschema.NewReferenceLoader("file:///brignext/schemas/secret.json"),
		service:             service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create project
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List projects
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Update project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.update),
	).Methods(http.MethodPut)

	// Delete project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// List secrets
	router.HandleFunc(
		"/v2/projects/{id}/secrets",
		e.TokenAuthFilter.Decorate(e.listSecrets),
	).Methods(http.MethodGet)

	// Set secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.setSecret),
	).Methods(http.MethodPut)

	// Unset secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.unsetSecret),
	).Methods(http.MethodDelete)

	// Grant role
	router.HandleFunc(
		"/v2/projects/{id}/role-assignments",
		e.TokenAuthFilter.Decorate(e.grantRole),
	).Methods(http.MethodPost)

	// Revoke role
	router.HandleFunc(
		"/v2/projects/{id}/role-assignments",
		e.TokenAuthFilter.Decorate(e.revokeRole),
	).Methods(http.MethodDelete)
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				return e.service.Create(r.Context(), project)
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
				return e.service.List(r.Context(), core.ProjectsSelector{}, opts)
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
		},
	)
}

func (e *endpoints) update(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				if mux.Vars(r)["id"] != project.ID {
					return nil, &core.ErrBadRequest{
						Reason: "The project IDs in the URL path and request body do " +
							"not match.",
					}
				}
				return e.service.Update(r.Context(), project)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) listSecrets(w http.ResponseWriter, r *http.Request) {
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
				return e.service.ListSecrets(r.Context(), mux.Vars(r)["id"], opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) setSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	secret := core.Secret{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: e.secretSchemaLoader,
			ReqBodyObj:          &secret,
			EndpointLogic: func() (interface{}, error) {
				if key != secret.Key {
					return nil, &core.ErrBadRequest{
						Reason: "The secret key in the URL path and request body do not " +
							"match.",
					}
				}
				return nil, e.service.SetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					secret,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) unsetSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.UnsetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					key,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) grantRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.RoleAssignment{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.GrantRole(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) revokeRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.RoleAssignment{
		Role:             r.URL.Query().Get("role"),
		UserID:           r.URL.Query().Get("userID"),
		ServiceAccountID: r.URL.Query().Get("serviceAccountID"),
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.RevokeRole(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
