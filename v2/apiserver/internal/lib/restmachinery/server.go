package restmachinery

import (
	"fmt"
	"log"
	"net/http"

	"github.com/brigadecore/brigade/v2/internal/file"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Server is an interface for the component that responds to HTTP API requests
type Server interface {
	// Run causes the API server to start serving HTTP requests. It will block
	// until an error occurs and will return that error.
	ListenAndServe() error
}

type server struct {
	*BaseEndpoints // The server itself exposes health check endpoints
	config         Config
	endpoints      []Endpoints
	handler        http.Handler
}

// NewServer returns a REST API server
func NewServer(
	config Config,
	baseEndpoints *BaseEndpoints,
	endpoints []Endpoints,
) Server {
	router := mux.NewRouter()
	router.StrictSlash(true)

	for _, eps := range endpoints {
		eps.Register(router)
	}

	s := &server{
		BaseEndpoints: baseEndpoints,
		config:        config,
		endpoints:     endpoints,
		handler: cors.New(
			cors.Options{
				AllowedMethods: []string{"DELETE", "GET", "POST", "PUT"},
			},
		).Handler(router),
	}

	// Health check
	router.HandleFunc(
		"/healthz",
		s.checkHealth, // No filters applied to this request
	).Methods(http.MethodGet)

	return s
}

func (s *server) ListenAndServe() error {
	address := fmt.Sprintf(":%d", s.config.Port())
	if s.config.TLSEnabled() &&
		file.Exists(s.config.TLSCertPath()) &&
		file.Exists(s.config.TLSKeyPath()) {
		log.Printf(
			"API server is listening with TLS enabled on 0.0.0.0:%d",
			s.config.Port(),
		)
		return http.ListenAndServeTLS(
			address,
			s.config.TLSCertPath(),
			s.config.TLSKeyPath(),
			s.handler,
		)
	}
	log.Printf(
		"API server is listening without TLS on 0.0.0.0:%d",
		s.config.Port(),
	)
	return http.ListenAndServe(
		address,
		h2c.NewHandler(s.handler, &http2.Server{}),
	)
}

// TODO: Develop a dedicated component to conduct healthchecks of database and
// message bus connections.
func (s *server) checkHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.ServeRequest(
		InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return struct{}{}, nil
			},
			SuccessCode: http.StatusOK,
		},
	)
}
