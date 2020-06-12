package main

import (
	"log"

	"github.com/krancour/brignext/v2/apiserver/internal/events"
	"github.com/krancour/brignext/v2/apiserver/internal/projects"
	"github.com/krancour/brignext/v2/apiserver/internal/serviceaccounts"
	"github.com/krancour/brignext/v2/apiserver/internal/sessions"
	"github.com/krancour/brignext/v2/apiserver/internal/users"
	"github.com/krancour/brignext/v2/internal/api"
	"github.com/krancour/brignext/v2/internal/api/auth"
	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/mongodb"
	"github.com/krancour/brignext/v2/internal/oidc"
	"github.com/krancour/brignext/v2/internal/redis"
	"github.com/krancour/brignext/v2/internal/version"
)

func main() {
	log.Printf(
		"Starting BrigNext API Server -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	// API server config
	apiConfig, err := api.GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	// Common
	database, err := mongodb.Database()
	if err != nil {
		log.Fatal(err)
	}
	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Events
	redisClient, err := redis.Client()
	if err != nil {
		log.Fatal(err)
	}
	eventsStore, err := events.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	eventsService := events.NewService(
		eventsStore,
		events.NewScheduler(redisClient, kubeClient),
		events.NewLogsStore(database),
	)

	// Projects
	projectsStore, err := projects.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	projectsService := projects.NewService(
		projectsStore,
		projects.NewScheduler(kubeClient),
	)

	// Service Accounts
	serviceAccountsStore, err := serviceaccounts.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	serviceAccountsService := serviceaccounts.NewService(serviceAccountsStore)

	// Sessions
	sessionsStore, err := sessions.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	sessionsService := sessions.NewService(sessionsStore)

	// Users
	usersStore, err := users.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	usersService := users.NewService(usersStore)

	baseEndpoints := &api.BaseEndpoints{
		TokenAuthFilter: auth.NewTokenAuthFilter(
			sessionsService.GetByToken,
			usersService.Get,
			apiConfig.RootUserEnabled(),
			apiConfig.HashedControllerToken(),
		),
	}

	// TODO: Move this
	// OpenID Connect config
	oidcConfig, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		api.NewServer(
			apiConfig,
			baseEndpoints,
			[]api.Endpoints{
				events.NewEndpoints(baseEndpoints, eventsService),
				projects.NewEndpoints(baseEndpoints, projectsService),
				serviceaccounts.NewEndpoints(baseEndpoints, serviceAccountsService),
				sessions.NewEndpoints(
					baseEndpoints,
					apiConfig.RootUserEnabled(),
					apiConfig.HashedRootUserPassword(),
					oidcConfig != nil && oidcIdentityVerifier != nil,
					oidcConfig,
					oidcIdentityVerifier,
					sessionsService,
					usersService,
				),
				users.NewEndpoints(baseEndpoints, usersService),
			},
		).ListenAndServe(),
	)
}
