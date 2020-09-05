package rest

import (
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type sessionsEndpoints struct {
	*restmachinery.BaseEndpoints
	service authx.SessionsService
}

func NewSessionsEndpoints(
	baseEndpoints *restmachinery.BaseEndpoints,
	service authx.SessionsService,
) restmachinery.Endpoints {
	return &sessionsEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (s *sessionsEndpoints) Register(router *mux.Router) {
	// Create session
	router.HandleFunc(
		"/v2/sessions",
		s.create, // No filters applied to this request
	).Methods(http.MethodPost)

	// Delete session
	router.HandleFunc(
		"/v2/session",
		s.TokenAuthFilter.Decorate(s.delete),
	).Methods(http.MethodDelete)

	// OIDC callback
	router.HandleFunc(
		"/v2/session/auth",
		s.authenticate, // No filters applied to this request
	).Methods(http.MethodGet)
}

func (s *sessionsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		s.ServeRequest(
			restmachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					username, password, ok := r.BasicAuth()
					if !ok {
						return nil, &meta.ErrBadRequest{
							Reason: "The request to create a new root session did not " +
								"include a valid basic auth header.",
						}
					}
					return s.service.CreateRootSession(r.Context(), username, password)
				},
				SuccessCode: http.StatusCreated,
			},
		)
		return
	}

	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return s.service.CreateUserSession(r.Context())
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (s *sessionsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	s.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				sessionID := authx.SessionIDFromContext(r.Context())
				if sessionID == "" {
					return nil, errors.New(
						"error: delete session request authenticated, but no session ID " +
							"found in request context",
					)
				}
				return nil, s.service.Delete(r.Context(), sessionID)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (s *sessionsEndpoints) authenticate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	oidcCode := r.URL.Query().Get("code")

	s.ServeHumanRequest(restmachinery.HumanRequest{
		W: w,
		EndpointLogic: func() (interface{}, error) {
			if oauth2State == "" || oidcCode == "" {
				return nil, &meta.ErrBadRequest{
					Reason: `The OpenID Connect authentication completion request ` +
						`lacked one or both of the "oauth2State" and "oidcCode" ` +
						`query parameters.`,
				}
			}
			if err := s.service.Authenticate(
				r.Context(),
				oauth2State,
				oidcCode,
			); err != nil {
				return nil,
					errors.Wrap(err, "error completing OpenID Connect authentication")
			}
			return []byte("You're now authenticated. You may resume using the CLI."),
				nil
		},
		SuccessCode: http.StatusOK,
	})
}
