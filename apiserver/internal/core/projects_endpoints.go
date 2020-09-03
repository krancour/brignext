package core

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/xeipuuv/gojsonschema"
)

type projectsEndpoints struct {
	*apimachinery.BaseEndpoints
	projectSchemaLoader gojsonschema.JSONLoader
	secretSchemaLoader  gojsonschema.JSONLoader
	service             ProjectsService
}

func NewProjectsEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service ProjectsService,
) apimachinery.Endpoints {
	// nolint: lll
	return &projectsEndpoints{
		BaseEndpoints:       baseEndpoints,
		projectSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		secretSchemaLoader:  gojsonschema.NewReferenceLoader("file:///brignext/schemas/secret.json"),
		service:             service,
	}
}

func (p *projectsEndpoints) Register(router *mux.Router) {
	// Create Project
	router.HandleFunc(
		"/v2/projects",
		p.TokenAuthFilter.Decorate(p.create),
	).Methods(http.MethodPost)

	// List Projects
	router.HandleFunc(
		"/v2/projects",
		p.TokenAuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Get Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.get),
	).Methods(http.MethodGet)

	// Update Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.update),
	).Methods(http.MethodPut)

	// Delete Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.TokenAuthFilter.Decorate(p.delete),
	).Methods(http.MethodDelete)

	// List Secrets
	router.HandleFunc(
		"/v2/projects/{id}/secrets",
		p.TokenAuthFilter.Decorate(p.listSecrets),
	).Methods(http.MethodGet)

	// Set Secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		p.TokenAuthFilter.Decorate(p.setSecret),
	).Methods(http.MethodPut)

	// Unset Secret
	router.HandleFunc(
		"/v2/projects/{id}/secrets/{key}",
		p.TokenAuthFilter.Decorate(p.unsetSecret),
	).Methods(http.MethodDelete)

	// Grant role to User
	router.HandleFunc(
		"/v2/projects/{id}/user-role-assignments",
		p.TokenAuthFilter.Decorate(p.grantUserRole),
	).Methods(http.MethodPost)

	// Revoke role from User
	router.HandleFunc(
		"/v2/projects/{id}/user-role-assignments",
		p.TokenAuthFilter.Decorate(p.revokeUserRole),
	).Methods(http.MethodDelete)

	// Grant role to ServiceAccount
	router.HandleFunc(
		"/v2/projects/{id}/service-account-role-assignments",
		p.TokenAuthFilter.Decorate(p.grantServiceAccountRole),
	).Methods(http.MethodPost)

	// Revoke role from ServiceAccount
	router.HandleFunc(
		"/v2/projects/{id}/service-account-role-assignments",
		p.TokenAuthFilter.Decorate(p.revokeServiceAccountRole),
	).Methods(http.MethodDelete)
}

func (p *projectsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	project := Project{}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				return p.service.Create(r.Context(), project)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *projectsEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			p.WriteAPIResponse(
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
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.service.List(r.Context(), ProjectsSelector{}, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) get(w http.ResponseWriter, r *http.Request) {
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) update(w http.ResponseWriter, r *http.Request) {
	project := Project{}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.projectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				if mux.Vars(r)["id"] != project.ID {
					return nil, &meta.ErrBadRequest{
						Reason: "The project IDs in the URL path and request body do " +
							"not match.",
					}
				}
				return p.service.Update(r.Context(), project)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) listSecrets(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			p.WriteAPIResponse(
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
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.service.ListSecrets(r.Context(), mux.Vars(r)["id"], opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) setSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	secret := Secret{}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.secretSchemaLoader,
			ReqBodyObj:          &secret,
			EndpointLogic: func() (interface{}, error) {
				if key != secret.Key {
					return nil, &meta.ErrBadRequest{
						Reason: "The secret key in the URL path and request body do not " +
							"match.",
					}
				}
				return nil, p.service.SetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					secret,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsEndpoints) unsetSecret(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.UnsetSecret(
					r.Context(),
					mux.Vars(r)["id"],
					key,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (p *projectsEndpoints) grantUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.GrantRoleToUser(
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

// TODO: This still needs some validation
func (p *projectsEndpoints) revokeUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{
		Role:   r.URL.Query().Get("role"),
		UserID: r.URL.Query().Get("userID"),
	}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.RevokeRoleFromUser(
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

// TODO: This still needs some validation
func (p *projectsEndpoints) grantServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request) {
	roleAssignment := authx.ServiceAccountRoleAssignment{}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.GrantRoleToServiceAccount(
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

// TODO: This still needs some validation
func (p *projectsEndpoints) revokeServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.ServiceAccountRoleAssignment{
		Role:             r.URL.Query().Get("role"),
		ServiceAccountID: r.URL.Query().Get("serviceAccountID"),
	}
	p.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.RevokeRoleFromServiceAccount(
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
