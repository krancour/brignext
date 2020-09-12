package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
)

type usersEndpoints struct {
	*restmachinery.BaseEndpoints
	service authx.UsersService
}

// TODO: There probably isn't any good reason to actually have this
// constructor-like function here. Let's consider removing it.
func NewUsersEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service authx.UsersService,
) restmachinery.Endpoints {
	return &usersEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (u *usersEndpoints) Register(router *mux.Router) {
	// List users
	router.HandleFunc(
		"/v2/users",
		u.TokenAuthFilter.Decorate(u.list),
	).Methods(http.MethodGet)

	// Get user
	router.HandleFunc(
		"/v2/users/{id}",
		u.TokenAuthFilter.Decorate(u.get),
	).Methods(http.MethodGet)

	// Lock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.TokenAuthFilter.Decorate(u.lock),
	).Methods(http.MethodPut)

	// Unlock user
	router.HandleFunc(
		"/v2/users/{id}/lock",
		u.TokenAuthFilter.Decorate(u.unlock),
	).Methods(http.MethodDelete)
}

func (u *usersEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			var err error
			if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
				opts.Limit < 1 || opts.Limit > 100 {
				u.WriteAPIResponse(
					w,
					http.StatusBadRequest,
					&meta.ErrBadRequest{
						Reason: fmt.Sprintf(
							`Invalid value %q for "limit" query parameter`,
							limitStr,
						),
					},
				)
			}
			return
		}
	}
	u.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return u.service.List(r.Context(), authx.UsersSelector{}, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersEndpoints) get(w http.ResponseWriter, r *http.Request) {
	u.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return u.service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersEndpoints) lock(w http.ResponseWriter, r *http.Request) {
	u.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, u.service.Lock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersEndpoints) unlock(w http.ResponseWriter, r *http.Request) {
	u.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, u.service.Unlock(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
