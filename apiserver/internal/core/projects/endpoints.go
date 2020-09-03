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
	// Create Project
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.create),
	).Methods(http.MethodPost)

	// List Projects
	router.HandleFunc(
		"/v2/projects",
		e.TokenAuthFilter.Decorate(e.list),
	).Methods(http.MethodGet)

	// Get Project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.get),
	).Methods(http.MethodGet)

	// Update Project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.update),
	).Methods(http.MethodPut)

	// Delete Project
	router.HandleFunc(
		"/v2/projects/{id}",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// List Secrets
	router.HandleFunc(
		"/v2/projects/{id}/secrets",
		e.TokenAuthFilter.Decorate(e.listSecrets),
	).Methods(http.MethodGet)

	// Set Secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.setSecret),
	).Methods(http.MethodPut)

	// Unset Secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		e.TokenAuthFilter.Decorate(e.unsetSecret),
	).Methods(http.MethodDelete)

	// Grant role to User
	router.HandleFunc(
		"/v2/projects/{id}/user-role-assignments",
		e.TokenAuthFilter.Decorate(e.grantUserRole),
	).Methods(http.MethodPost)

	// Revoke role from User
	router.HandleFunc(
		"/v2/projects/{id}/user-role-assignments",
		e.TokenAuthFilter.Decorate(e.revokeUserRole),
	).Methods(http.MethodDelete)

	// Grant role to ServiceAccount
	router.HandleFunc(
		"/v2/projects/{id}/service-account-role-assignments",
		e.TokenAuthFilter.Decorate(e.grantServiceAccountRole),
	).Methods(http.MethodPost)

	// Revoke role from ServiceAccount
	router.HandleFunc(
		"/v2/projects/{id}/service-account-role-assignments",
		e.TokenAuthFilter.Decorate(e.revokeServiceAccountRole),
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
					return nil, &meta.ErrBadRequest{
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
					return nil, &meta.ErrBadRequest{
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

func (e *endpoints) grantUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.GrantRoleToUser(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) revokeUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{
		Role:   r.URL.Query().Get("role"),
		UserID: r.URL.Query().Get("userID"),
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.RevokeRoleFromUser(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) grantServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request) {
	roleAssignment := authx.ServiceAccountRoleAssignment{}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.GrantRoleToServiceAccount(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) revokeServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.ServiceAccountRoleAssignment{
		Role:             r.URL.Query().Get("role"),
		ServiceAccountID: r.URL.Query().Get("serviceAccountID"),
	}
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, e.service.RevokeRoleFromServiceAccount(
					r.Context(),
					mux.Vars(r)["id"],
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
