package system

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	service Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	// nolint: lll
	return &endpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Grant role to User
	router.HandleFunc(
		"/v2/system/user-role-assignments",
		e.TokenAuthFilter.Decorate(e.grantUserRole),
	).Methods(http.MethodPost)

	// Revoke role from User
	router.HandleFunc(
		"/v2/system/user-role-assignments",
		e.TokenAuthFilter.Decorate(e.revokeUserRole),
	).Methods(http.MethodDelete)

	// Grant role to ServiceAccount
	router.HandleFunc(
		"/v2/system/service-account-role-assignments",
		e.TokenAuthFilter.Decorate(e.grantServiceAccountRole),
	).Methods(http.MethodPost)

	// Revoke role from ServiceAccount
	router.HandleFunc(
		"/v2/system/service-account-role-assignments",
		e.TokenAuthFilter.Decorate(e.revokeServiceAccountRole),
	).Methods(http.MethodDelete)
}

// TODO: This still needs some validation
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
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
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
					roleAssignment.UserID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
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
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

// TODO: This still needs some validation
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
					roleAssignment.ServiceAccountID,
					roleAssignment.Role,
				)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
