package api

import (
	"fmt"
	"log"
	"net/http"

	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/file"
	"github.com/krancour/brignext/pkg/http/filters/auth"
	"github.com/krancour/brignext/pkg/storage"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/oauth2"
)

// Server is an interface for the component that responds to HTTP API requests
type Server interface {
	// Run causes the API server to start serving HTTP requests. It will block
	// until an error occurs and will return that error.
	ListenAndServe() error
}

type server struct {
	apiServerConfig   Config
	oauth2Config      *oauth2.Config
	oidcTokenVerifier *oidc.IDTokenVerifier
	userStore         storage.UserStore
	sessionStore      storage.SessionStore
	oldProjectStore   oldStorage.Store
	projectStore      storage.ProjectStore
	logStore          storage.LogStore
	router            *mux.Router
}

// NewServer returns an HTTP router
func NewServer(
	apiServerConfig Config,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
	userStore storage.UserStore,
	sessionStore storage.SessionStore,
	oldProjectStore oldStorage.Store,
	projectStore storage.ProjectStore,
	logStore storage.LogStore,
) Server {
	s := &server{
		apiServerConfig:   apiServerConfig,
		oauth2Config:      oauth2Config,
		oidcTokenVerifier: oidcTokenVerifier,
		userStore:         userStore,
		sessionStore:      sessionStore,
		oldProjectStore:   oldProjectStore,
		projectStore:      projectStore,
		logStore:          logStore,
		router:            mux.NewRouter(),
	}

	// Most requests are authenticated with a bearer token
	tokenAuthFilter := auth.NewTokenAuthFilter(
		sessionStore.GetSessionByToken,
		userStore.GetUser,
		apiServerConfig.RootUserEnabled(),
	)

	s.router.StrictSlash(true)

	// Create session
	s.router.HandleFunc(
		"/v2/sessions",
		s.sessionCreate, // No filters applied to this request
	).Methods(http.MethodPost)

	if oauth2Config != nil && oidcTokenVerifier != nil {
		// OIDC callback
		s.router.HandleFunc(
			"/auth/oidc/callback",
			s.oidcAuthComplete, // No filters applied to this request
		).Methods(http.MethodGet)
	}

	// Delete session
	s.router.HandleFunc(
		"/v2/session",
		tokenAuthFilter.Decorate(s.sessionDelete),
	).Methods(http.MethodDelete)

	// Create project
	s.router.HandleFunc(
		"/v2/projects",
		tokenAuthFilter.Decorate(s.projectCreate),
	).Methods(http.MethodPost)

	// List projects
	s.router.HandleFunc(
		"/v2/projects",
		tokenAuthFilter.Decorate(s.projectList),
	).Methods(http.MethodGet)

	// Get project
	s.router.HandleFunc(
		"/v2/projects/{name}",
		tokenAuthFilter.Decorate(s.projectGet),
	).Methods(http.MethodGet)

	// List project's builds
	s.router.HandleFunc(
		"/v2/projects/{projectName}/builds",
		tokenAuthFilter.Decorate(s.buildList),
	).Methods(http.MethodGet)

	// Update project
	s.router.HandleFunc(
		"/v2/projects/{name}",
		tokenAuthFilter.Decorate(s.projectUpdate),
	).Methods(http.MethodPut)

	// Delete project
	s.router.HandleFunc(
		"/v2/projects/{name}",
		tokenAuthFilter.Decorate(s.projectDelete),
	).Methods(http.MethodDelete)

	// List project's builds
	s.router.HandleFunc(
		"/v2/projects/{projectName}/builds",
		tokenAuthFilter.Decorate(s.buildDeleteAll),
	).Methods(http.MethodDelete)

	// Create build
	s.router.HandleFunc(
		"/v2/builds",
		tokenAuthFilter.Decorate(s.buildCreate),
	).Methods(http.MethodPost)

	// List builds
	s.router.HandleFunc(
		"/v2/builds",
		tokenAuthFilter.Decorate(s.buildList),
	).Methods(http.MethodGet)

	// Get build
	s.router.HandleFunc(
		"/v2/builds/{id}",
		tokenAuthFilter.Decorate(s.buildGet),
	).Methods(http.MethodGet)

	// Stream logs
	s.router.HandleFunc(
		"/v2/builds/{id}/logs",
		tokenAuthFilter.Decorate(s.buildLogs),
	).Methods(http.MethodGet)

	// Delete build
	s.router.HandleFunc(
		"/v2/builds/{id}",
		tokenAuthFilter.Decorate(s.buildDelete),
	).Methods(http.MethodDelete)

	// Health check
	s.router.HandleFunc(
		"/healthz",
		s.healthCheck, // No filters applied to this request
	).Methods(http.MethodGet)

	return s
}

func (s *server) ListenAndServe() error {
	address := fmt.Sprintf(":%d", s.apiServerConfig.Port())
	if s.apiServerConfig.TLSEnabled() &&
		file.Exists(s.apiServerConfig.TLSCertPath()) &&
		file.Exists(s.apiServerConfig.TLSKeyPath()) {
		log.Printf(
			"API server is listening with TLS enabled on 0.0.0.0:%d",
			s.apiServerConfig.Port(),
		)
		return http.ListenAndServeTLS(
			address,
			s.apiServerConfig.TLSCertPath(),
			s.apiServerConfig.TLSKeyPath(),
			s.router,
		)
	}
	log.Printf(
		"API server is listening without TLS on 0.0.0.0:%d",
		s.apiServerConfig.Port(),
	)
	return http.ListenAndServe(
		address,
		h2c.NewHandler(s.router, &http2.Server{}),
	)
}
