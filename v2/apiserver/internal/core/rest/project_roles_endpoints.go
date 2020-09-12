package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
)

type projectsRolesEndpoints struct {
	*restmachinery.BaseEndpoints
	service core.ProjectRolesService
}

// TODO: There probably isn't any good reason to actually have this
// constructor-like function here. Let's consider removing it.
func NewProjectsRolesEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service core.ProjectRolesService,
) restmachinery.Endpoints {
	// nolint: lll
	return &projectsRolesEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (p *projectsRolesEndpoints) Register(router *mux.Router) {
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

// TODO: This still needs some validation via JSON schema
func (p *projectsRolesEndpoints) grant(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.RoleAssignment{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.Grant(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation via JSON schema
func (p *projectsRolesEndpoints) revoke(
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
				return nil, p.service.Revoke(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
