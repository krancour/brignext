package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	"github.com/gorilla/mux"
)

type RolesEndpoints struct {
	*restmachinery.BaseEndpoints
	Service system.RolesService
}

func (r *RolesEndpoints) Register(router *mux.Router) {
	// Grant a system Role to a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		r.TokenAuthFilter.Decorate(r.grant),
	).Methods(http.MethodPost)

	// Revoke a system Role for a User or ServiceAccount
	router.HandleFunc(
		"/v2/system/role-assignments",
		r.TokenAuthFilter.Decorate(r.revoke),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation via JSON schema
func (r *RolesEndpoints) grant(
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
				return nil, r.Service.Grant(
					req.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation via JSON schema
func (r *RolesEndpoints) revoke(
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
				return nil, r.Service.Revoke(
					req.Context(),
					roleAssignment,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
