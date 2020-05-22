package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
)

func (s *server) sessionCreate(w http.ResponseWriter, r *http.Request) {
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
				return s.service.Sessions().CreateRootSession(r.Context())
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
			oauth2State, token, err :=
				s.service.Sessions().CreateUserSession(r.Context())
			if err != nil {
				return nil, err
			}
			// TODO: This should be a more formalized type... but what to call it...
			return struct {
				Token   string `json:"token"`
				AuthURL string `json:"authURL"`
			}{
				Token:   token,
				AuthURL: s.oauth2Config.AuthCodeURL(oauth2State),
			}, nil
		},
		successCode: http.StatusCreated,
	})
}

func (s *server) sessionDelete(w http.ResponseWriter, r *http.Request) {
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
			return nil, s.service.Sessions().Delete(r.Context(), sessionID)
		},
		successCode: http.StatusOK,
	})
}
