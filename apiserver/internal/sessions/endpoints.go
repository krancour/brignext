package sessions

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/auth"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
)

type endpoints struct {
	*apimachinery.BaseEndpoints
	service Service
}

func NewEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service Service,
) apimachinery.Endpoints {
	return &endpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (e *endpoints) Register(router *mux.Router) {
	// Create session
	router.HandleFunc(
		"/v2/sessions",
		e.create, // No filters applied to this request
	).Methods(http.MethodPost)

	// Delete session
	router.HandleFunc(
		"/v2/session",
		e.TokenAuthFilter.Decorate(e.delete),
	).Methods(http.MethodDelete)

	// OIDC callback
	router.HandleFunc(
		"/v2/session/auth",
		e.authenticate, // No filters applied to this request
	).Methods(http.MethodGet)
}

func (e *endpoints) create(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		e.ServeRequest(
			apimachinery.InboundRequest{
				W: w,
				R: r,
				EndpointLogic: func() (interface{}, error) {
					username, password, ok := r.BasicAuth()
					if !ok {
						return nil, &brignext.ErrBadRequest{
							Reason: "The request to create a new root session did not " +
								"include a valid basic auth header.",
						}
					}
					return e.service.CreateRootSession(r.Context(), username, password)
				},
				SuccessCode: http.StatusCreated,
			},
		)
		return
	}

	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return e.service.CreateUserSession(r.Context())
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (e *endpoints) delete(w http.ResponseWriter, r *http.Request) {
	e.ServeRequest(
		apimachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				sessionID := auth.SessionIDFromContext(r.Context())
				if sessionID == "" {
					return nil, errors.New(
						"error: delete session request authenticated, but no session ID " +
							"found in request context",
					)
				}
				return nil, e.service.Delete(r.Context(), sessionID)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (e *endpoints) authenticate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	oauth2State := r.URL.Query().Get("state")
	oidcCode := r.URL.Query().Get("code")

	e.ServeHumanRequest(apimachinery.HumanRequest{
		W: w,
		EndpointLogic: func() (interface{}, error) {
			if oauth2State == "" || oidcCode == "" {
				return nil, &brignext.ErrBadRequest{
					Reason: `The OpenID Connect authentication completion request ` +
						`lacked one or both of the "oauth2State" and "oidcCode" ` +
						`query parameters.`,
				}
			}
			if err := e.service.Authenticate(
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
