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
	// Grant role to User
	router.HandleFunc(
		"/v2/system/user-role-assignments",
		s.TokenAuthFilter.Decorate(s.grantUserRole),
	).Methods(http.MethodPost)

	// Revoke role from User
	router.HandleFunc(
		"/v2/system/user-role-assignments",
		s.TokenAuthFilter.Decorate(s.revokeUserRole),
	).Methods(http.MethodDelete)

	// Grant role to ServiceAccount
	router.HandleFunc(
		"/v2/system/service-account-role-assignments",
		s.TokenAuthFilter.Decorate(s.grantServiceAccountRole),
	).Methods(http.MethodPost)

	// Revoke role from ServiceAccount
	router.HandleFunc(
		"/v2/system/service-account-role-assignments",
		s.TokenAuthFilter.Decorate(s.revokeServiceAccountRole),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) grantUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.GrantToUser(
					r.Context(),
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) revokeUserRole(w http.ResponseWriter, r *http.Request) {
	roleAssignment := authx.UserRoleAssignment{
		Role:   r.URL.Query().Get("role"),
		UserID: r.URL.Query().Get("userID"),
	}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.RevokeFromUser(
					r.Context(),
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) grantServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request) {
	roleAssignment := authx.ServiceAccountRoleAssignment{}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W:          w,
			R:          r,
			ReqBodyObj: &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.GrantToServiceAccount(
					r.Context(),
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
func (s *systemRolesEndpoints) revokeServiceAccountRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := authx.ServiceAccountRoleAssignment{
		Role:             r.URL.Query().Get("role"),
		ServiceAccountID: r.URL.Query().Get("serviceAccountID"),
	}
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, s.service.RevokeFromServiceAccount(
					r.Context(),
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
