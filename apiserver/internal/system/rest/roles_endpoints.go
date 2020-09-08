package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/gorilla/mux"
)

type rolesEndpoints struct {
	*restmachinery.BaseEndpoints
	service system.RolesService
}

func NewRolesEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service system.RolesService,
) restmachinery.Endpoints {
	// nolint: lll
	return &rolesEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (r *rolesEndpoints) Register(router *mux.Router) {
	// Grant a system Role to a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		r.TokenAuthFilter.Decorate(r.grantRole),
	).Methods(http.MethodPost)

	// Revoke a system Role for a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		r.TokenAuthFilter.Decorate(r.revokeRole),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation
func (r *rolesEndpoints) grantRole(
	w http.ResponseWriter,
	req *http.Request,
) {
	roleAssignment := authx.RoleAssignment{}
	r.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          req,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, r.service.GrantRole(
					req.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (r *rolesEndpoints) revokeRole(
	w http.ResponseWriter,
	req *http.Request,
) {
	roleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName(req.URL.Query().Get("role")),
		PrincipalType: authx.PrincipalType(req.URL.Query().Get("principalType")),
		PrincipalID:   req.URL.Query().Get("principalID"),
	}
	r.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: req,
			EndpointLogic: func() (interface{}, error) {
				return nil, r.service.RevokeRole(
					req.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}