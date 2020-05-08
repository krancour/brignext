package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/pkg/api/auth"
	"github.com/krancour/brignext/v2/apiserver/pkg/service"
	"github.com/krancour/brignext/v2/pkg/file"
	"github.com/xeipuuv/gojsonschema"
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
	apiServerConfig            Config
	oauth2Config               *oauth2.Config
	oidcTokenVerifier          *oidc.IDTokenVerifier
	service                    service.Service
	router                     *mux.Router
	serviceAccountSchemaLoader gojsonschema.JSONLoader
	projectSchemaLoader        gojsonschema.JSONLoader
	eventSchemaLoader          gojsonschema.JSONLoader
	workerStatusSchemaLoader   gojsonschema.JSONLoader
	jobStatusSchemaLoader      gojsonschema.JSONLoader
}

// NewServer returns an HTTP router
func NewServer(
	apiServerConfig Config,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
	service service.Service,
) Server {
	// nolint: lll
	s := &server{
		apiServerConfig:            apiServerConfig,
		oauth2Config:               oauth2Config,
		oidcTokenVerifier:          oidcTokenVerifier,
		service:                    service,
		router:                     mux.NewRouter(),
		serviceAccountSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/service-account.json"),
		projectSchemaLoader:        gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		eventSchemaLoader:          gojsonschema.NewReferenceLoader("file:///brignext/schemas/event.json"),
		workerStatusSchemaLoader:   gojsonschema.NewReferenceLoader("file:///brignext/schemas/worker-status.json"),
		jobStatusSchemaLoader:      gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-status.json"),
	}

	// Most requests are authenticated with a bearer token
	tokenAuthFilter := auth.NewTokenAuthFilter(
		service.GetSessionByToken,
		service.GetUser,
		apiServerConfig.RootUserEnabled(),
		apiServerConfig.HashedControllerToken(),
	)

	s.router.StrictSlash(true)

	// List users
	s.router.HandleFunc(
		"/v2/users",
		tokenAuthFilter.Decorate(s.userList),
	).Methods(http.MethodGet)

	// Get user
	s.router.HandleFunc(
		"/v2/users/{id}",
		tokenAuthFilter.Decorate(s.userGet),
	).Methods(http.MethodGet)

	// Lock user
	s.router.HandleFunc(
		"/v2/users/{id}/lock",
		tokenAuthFilter.Decorate(s.userLock),
	).Methods(http.MethodPost)

	// Unlock user
	s.router.HandleFunc(
		"/v2/users/{id}/lock",
		tokenAuthFilter.Decorate(s.userUnlock),
	).Methods(http.MethodDelete)

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

	// Create service account
	s.router.HandleFunc(
		"/v2/service-accounts",
		tokenAuthFilter.Decorate(s.serviceAccountCreate),
	).Methods(http.MethodPost)

	// List service accounts
	s.router.HandleFunc(
		"/v2/service-accounts",
		tokenAuthFilter.Decorate(s.serviceAccountList),
	).Methods(http.MethodGet)

	// Get service account
	s.router.HandleFunc(
		"/v2/service-accounts/{id}",
		tokenAuthFilter.Decorate(s.serviceAccountGet),
	).Methods(http.MethodGet)

	// Lock service account
	s.router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		tokenAuthFilter.Decorate(s.serviceAccountLock),
	).Methods(http.MethodPost)

	// Unlock service account
	s.router.HandleFunc(
		"/v2/service-accounts/{id}/lock",
		tokenAuthFilter.Decorate(s.serviceAccountUnlock),
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
		"/v2/projects/{id}",
		tokenAuthFilter.Decorate(s.projectGet),
	).Methods(http.MethodGet)

	// Update project
	s.router.HandleFunc(
		"/v2/projects/{id}",
		tokenAuthFilter.Decorate(s.projectUpdate),
	).Methods(http.MethodPut)

	// Delete project
	s.router.HandleFunc(
		"/v2/projects/{id}",
		tokenAuthFilter.Decorate(s.projectDelete),
	).Methods(http.MethodDelete)

	// List secrets
	s.router.HandleFunc(
		"/v2/projects/{projectID}/worker/secrets",
		tokenAuthFilter.Decorate(s.secretsList),
	).Methods(http.MethodGet)

	// Set secrets
	s.router.HandleFunc(
		"/v2/projects/{projectID}/worker/secrets",
		tokenAuthFilter.Decorate(s.secretsSet),
	).Methods(http.MethodPost)

	// Unset secrets
	s.router.HandleFunc(
		"/v2/projects/{projectID}/worker/secrets",
		tokenAuthFilter.Decorate(s.secretsUnset),
	).Methods(http.MethodDelete)

	// Create event
	s.router.HandleFunc(
		"/v2/events",
		tokenAuthFilter.Decorate(s.eventCreate),
	).Methods(http.MethodPost)

	// List events
	s.router.HandleFunc(
		"/v2/events",
		tokenAuthFilter.Decorate(s.eventList),
	).Methods(http.MethodGet)

	// Get event
	s.router.HandleFunc(
		"/v2/events/{id}",
		tokenAuthFilter.Decorate(s.eventGet),
	).Methods(http.MethodGet)

	// Cancel event
	s.router.HandleFunc(
		"/v2/events/{id}/cancel",
		tokenAuthFilter.Decorate(s.eventsCancel),
	).Methods(http.MethodPut)

	// Cancel events by project
	s.router.HandleFunc(
		"/v2/projects/{projectID}/events/cancel",
		tokenAuthFilter.Decorate(s.eventsCancel),
	).Methods(http.MethodPut)

	// Delete event
	s.router.HandleFunc(
		"/v2/events/{id}",
		tokenAuthFilter.Decorate(s.eventsDelete),
	).Methods(http.MethodDelete)

	// Delete events by project
	s.router.HandleFunc(
		"/v2/projects/{projectID}/events",
		tokenAuthFilter.Decorate(s.eventsDelete),
	).Methods(http.MethodDelete)

	// Update worker status
	s.router.HandleFunc(
		"/v2/events/{eventID}/worker/status",
		tokenAuthFilter.Decorate(s.workerUpdateStatus),
	).Methods(http.MethodPut)

	// Get/stream worker logs
	s.router.HandleFunc(
		"/v2/events/{eventID}/worker/logs",
		tokenAuthFilter.Decorate(s.workerLogs),
	).Methods(http.MethodGet)

	// Get job
	s.router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}",
		tokenAuthFilter.Decorate(s.jobGet),
	).Methods(http.MethodGet)

	// Update job status
	s.router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/status",
		tokenAuthFilter.Decorate(s.jobUpdateStatus),
	).Methods(http.MethodPut)

	// Get/stream job logs
	s.router.HandleFunc(
		"/v2/events/{eventID}/worker/jobs/{jobName}/logs",
		tokenAuthFilter.Decorate(s.jobLogs),
	).Methods(http.MethodGet)

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
