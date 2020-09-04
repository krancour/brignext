package rest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery"
)

type projectsRolesEndpoints struct {
	*restmachinery.BaseEndpoints
	service core.ProjectsService
}

func NewProjectsRolesEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service core.ProjectsService,
) restmachinery.Endpoints {
	// nolint: lll
	return &projectsEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (p *projectsRolesEndpoints) Register(router *mux.Router) {
	// Grant role to User
	router.HandleFunc(
		"/v2/projects/{projectID}/user-role-assignments",
		p.TokenAuthFilter.Decorate(p.grantToUser),
	).Methods(http.MethodPost)

	// Revoke role from User
	router.HandleFunc(
		"/v2/projects/{projectID}/user-role-assignments",
		p.TokenAuthFilter.Decorate(p.revokeFromUser),
	).Methods(http.MethodDelete)

	// Grant role to ServiceAccount
	router.HandleFunc(
		"/v2/projects/{projectID}/service-account-role-assignments",
		p.TokenAuthFilter.Decorate(p.grantToServiceAccount),
	).Methods(http.MethodPost)

	// Revoke role from ServiceAccount
	router.HandleFunc(
		"/v2/projects/{projectID}/service-account-role-assignments",
		p.TokenAuthFilter.Decorate(p.revokeFromServiceAccount),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation
func (p *projectsRolesEndpoints) grantToUser(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.GrantRoleToUser(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (p *projectsRolesEndpoints) revokeFromUser(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{
		Role:   r.URL.Query().Get("role"),
		UserID: r.URL.Query().Get("userID"),
	}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.RevokeRoleFromUser(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (p *projectsRolesEndpoints) grantToServiceAccount(
	w http.ResponseWriter,
	r *http.Request) {
	roleAssignment := authx.ServiceAccountRoleAssignment{}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.GrantRoleToServiceAccount(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (p *projectsRolesEndpoints) revokeFromServiceAccount(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.ServiceAccountRoleAssignment{
		Role:             r.URL.Query().Get("role"),
		ServiceAccountID: r.URL.Query().Get("serviceAccountID"),
	}
	p.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.service.RevokeRoleFromServiceAccount(
					r.Context(),
					mux.Vars(r)["projectID"],
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
