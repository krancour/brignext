package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/gorilla/mux"
)

type systemRolesEndpoints struct {
	*restmachinery.BaseEndpoints
	service system.SystemRolesService
}

func NewSystemRolesEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service system.SystemRolesService,
) restmachinery.Endpoints {
	// nolint: lll
	return &systemRolesEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (s *systemRolesEndpoints) Register(router *mux.Router) {
	// Grant a system Role to a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		s.TokenAuthFilter.Decorate(s.grantRole),
	).Methods(http.MethodPost)

	// Revoke a system Role for a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		s.TokenAuthFilter.Decorate(s.revokeRole),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) grantRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.RoleAssignment{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.GrantRole(
					r.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) revokeRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.RoleAssignment{
		Role:          authx.RoleName(r.URL.Query().Get("role")),
		PrincipalType: authx.PrincipalType(r.URL.Query().Get("principalType")),
		PrincipalID:   r.URL.Query().Get("principalID"),
	}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.RevokeRole(
					r.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
