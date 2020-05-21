package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/pkg/errors"
)

func (s *server) sessionCreate(w http.ResponseWriter, r *http.Request) {
	// nolint: errcheck
	rootSessionRequest, _ := strconv.ParseBool(r.URL.Query().Get("root"))

	if rootSessionRequest {
		if !s.apiServerConfig.RootUserEnabled() {
			s.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication(
					"authentication using root credentials is not supported by this "+
						"server",
				),
			)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			s.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication(
					"the request did not include a valid basic auth header",
				),
			)
			return
		}
		if username != "root" ||
			crypto.ShortSHA(username, password) !=
				s.apiServerConfig.HashedRootUserPassword() {
			s.writeResponse(
				w,
				http.StatusUnauthorized,
				brignext.NewErrAuthentication(
					"could not authenticate request using the supplied credentials",
				),
			)
			return
		}
		s.serveAPIRequest(apiRequest{
			w: w,
			r: r,
			endpointLogic: func() (interface{}, error) {
				return s.service.Sessions().CreateRootSession(r.Context())
			},
			successCode: http.StatusCreated,
		})
		return
	}

	if s.oauth2Config == nil || s.oidcTokenVerifier == nil {
		s.writeResponse(
			w,
			http.StatusUnauthorized,
			brignext.NewErrAuthentication(
				"authentication using OpenID Connect is not supported by this server",
			),
		)
		return
	}
	oauth2State, token, err := s.service.Sessions().CreateUserSession(r.Context())
	if err != nil {
		log.Println(
			errors.Wrap(err, "error creating new session"),
		)
		s.writeResponse(
			w,
			http.StatusInternalServerError,
			brignext.NewErrInternalServer(),
		)
		return
	}
	respStruct := struct {
		Token   string `json:"token"`
		AuthURL string `json:"authURL"`
	}{
		Token:   token,
		AuthURL: s.oauth2Config.AuthCodeURL(oauth2State),
	}
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return respStruct, nil
		},
		successCode: http.StatusCreated,
	})
}

func (s *server) sessionDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := auth.SessionIDFromContext(r.Context())
	if sessionID == "" {
		log.Println(
			"error: delete session request authenticated, but no session ID found " +
				"in request context",
		)
		s.writeResponse(
			w,
			http.StatusInternalServerError,
			brignext.NewErrInternalServer(),
		)
		return
	}
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			return nil, s.service.Sessions().Delete(r.Context(), sessionID)
		},
		successCode: http.StatusOK,
	})
}
