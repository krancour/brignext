package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/krancour/brignext/v2/internal/common/file"
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
	apiServerConfig Config
	router          *mux.Router
}

// NewServer returns an HTTP router
func NewServer(
	apiServerConfig Config,
	oauth2Config *oauth2.Config,
	oidcTokenVerifier *oidc.IDTokenVerifier,
	service service.Service,
) Server {

	router := mux.NewRouter()
	router.StrictSlash(true)

	baseEndpoints := &baseEndpoints{
		tokenAuthFilter: auth.NewTokenAuthFilter(
			service.Sessions().GetByToken,
			service.Users().Get,
			apiServerConfig.RootUserEnabled(),
			apiServerConfig.HashedControllerToken(),
		),
	}

	// nolint: lll
	(&serviceAccountEndpoints{
		baseEndpoints:              baseEndpoints,
		serviceAccountSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/service-account.json"),
		service:                    service.ServiceAccounts(),
	}).register(router)

	// nolint: lll
	(&eventEndpoints{
		baseEndpoints:            baseEndpoints,
		eventSchemaLoader:        gojsonschema.NewReferenceLoader("file:///brignext/schemas/event.json"),
		workerStatusSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/worker-status.json"),
		jobStatusSchemaLoader:    gojsonschema.NewReferenceLoader("file:///brignext/schemas/job-status.json"),
		service:                  service.Events(),
	}).register(router)

	(&userEndpoints{
		baseEndpoints: baseEndpoints,
		service:       service.Users(),
	}).register(router)

	(&sessionEndpoints{
		baseEndpoints:     baseEndpoints,
		apiServerConfig:   apiServerConfig,
		oauth2Config:      oauth2Config,
		oidcTokenVerifier: oidcTokenVerifier,
		service:           service.Sessions(),
	}).register(router)

	// nolint: lll
	(&projectEndpoints{
		baseEndpoints:       baseEndpoints,
		projectSchemaLoader: gojsonschema.NewReferenceLoader("file:///brignext/schemas/project.json"),
		secretSchemaLoader:  gojsonschema.NewReferenceLoader("file:///brignext/schemas/secret.json"),
		service:             service.Projects(),
	}).register(router)

	(&healthEndpoints{
		baseEndpoints: baseEndpoints,
	}).register(router)

	if oauth2Config != nil && oidcTokenVerifier != nil {
		(&oidcEndpoints{
			baseEndpoints:     baseEndpoints,
			oauth2Config:      oauth2Config,
			oidcTokenVerifier: oidcTokenVerifier,
			service:           service,
		}).register(router)
	}

	return &server{
		apiServerConfig: apiServerConfig,
		router:          router,
	}
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
