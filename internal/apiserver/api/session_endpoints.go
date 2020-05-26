package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"golang.org/x/oauth2"
)

type sessionEndpoints struct {
	*baseEndpoints
	apiServerConfig   Config
	oauth2Config      *oauth2.Config
	oidcTokenVerifier *oidc.IDTokenVerifier
	service           service.SessionsService
}

func (s *sessionEndpoints) register(router *mux.Router) {
	// Create session
	router.HandleFunc(
		"/v2/sessions",
		s.create, // No filters applied to this request
	).Methods(http.MethodPost)

	// Delete session
	router.HandleFunc(
		"/v2/session",
		s.tokenAuthFilter.Decorate(s.delete),
	).Methods(http.MethodDelete)
}

func (s *sessionEndpoints) create(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		s.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				if !s.apiServerConfig.RootUserEnabled() {
					return nil, brignext.NewErrNotSupported(
						"Authentication using root credentials is not supported by this " +
							"server.",
					)
				}
				username, password, ok := r.BasicAuth()
				if !ok {
					return nil, brignext.NewErrBadRequest(
						"The request to create a new root session did not include a " +
							"valid basic auth header.",
					)
				}
				if username != "root" ||
					crypto.ShortSHA(username, password) !=
						s.apiServerConfig.HashedRootUserPassword() {
					return nil, brignext.NewErrAuthentication(
						"Could not authenticate request using the supplied credentials.",
					)
				}
				return s.service.CreateRootSession(r.Context())
			},
			successCode: http.StatusCreated,
		})
		return
	}

	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			if s.oauth2Config == nil || s.oidcTokenVerifier == nil {
				return nil, brignext.NewErrNotSupported(
					"Authentication using OpenID Connect is not supported by this " +
						"server.",
				)
			}
			userSessionAuthDetails, err := s.service.CreateUserSession(r.Context())
			if err != nil {
				return nil, err
			}
			userSessionAuthDetails.AuthURL = s.oauth2Config.AuthCodeURL(
				userSessionAuthDetails.OAuth2State,
			)
			return userSessionAuthDetails, nil
		},
		successCode: http.StatusCreated,
	})
}

func (s *sessionEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			sessionID := auth.SessionIDFromContext(r.Context())
			if sessionID == "" {
				return nil, errors.New(
					"error: delete session request authenticated, but no session ID " +
						"found in request context",
				)
			}
			return nil, s.service.Delete(r.Context(), sessionID)
		},
		successCode: http.StatusOK,
	})
}
