package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type ProjectsRolesEndpoints struct {
	*restmachinery.BaseEndpoints
	ProjectRoleAssignmentSchemaLoader gojsonschema.JSONLoader
	Service                           core.ProjectRolesService
}

func (p *ProjectsRolesEndpoints) Register(router *mux.Router) {
	// Grant a Project Role to a User or Service Account
	router.HandleFunc(
		"/v2/projects/{projectID}/role-assignments",
		p.TokenAuthFilter.Decorate(p.grant),
	).Methods(http.MethodPost)

	// Revoke a Project Role for a User or Service Account
	router.HandleFunc(
		"/v2/projects/{projectID}/role-assignments",
		p.TokenAuthFilter.Decorate(p.revoke),
	).Methods(http.MethodDelete)
}

func (p *ProjectsRolesEndpoints) grant(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.RoleAssignment{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.ProjectRoleAssignmentSchemaLoader,
			ReqBodyObj:          &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Grant(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectsRolesEndpoints) revoke(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName(r.URL.Query().Get("role")),
		PrincipalType: authx.PrincipalType(r.URL.Query().Get("principalType")),
		PrincipalID:   r.URL.Query().Get("principalID"),
	}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Revoke(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
