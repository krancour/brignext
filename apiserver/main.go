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
	"github.com/krancour/brignext/v2/internal/events/amqp"
	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/mongodb"
	"github.com/krancour/brignext/v2/internal/oidc"
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

	// Projects
	projectsStore, err := projects.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	projectsService := projects.NewService(
		projectsStore,
		projects.NewScheduler(kubeClient),
	)

	// Events-- depends on projects
	eventSenderFactory, err := amqp.GetSenderFactoryFromEnvironment()
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
		projectsStore,
		eventsStore,
		events.NewLogsStore(database),
		events.NewScheduler(eventSenderFactory, kubeClient),
	)

	// Service Accounts
	serviceAccountsStore, err := serviceaccounts.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	serviceAccountsService := serviceaccounts.NewService(serviceAccountsStore)

	// Users
	usersStore, err := users.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	usersService := users.NewService(usersStore)

	// Sessions-- depends on users
	oauth2Config, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	sessionsStore, err := sessions.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	sessionsService := sessions.NewService(
		sessionsStore,
		usersStore,
		apiConfig.RootUserEnabled(),
		apiConfig.HashedRootUserPassword(),
		oauth2Config,
		oidcIdentityVerifier,
	)

	baseEndpoints := &api.BaseEndpoints{
		TokenAuthFilter: auth.NewTokenAuthFilter(
			sessionsService.GetByToken,
			usersService.Get,
			apiConfig.RootUserEnabled(),
			apiConfig.HashedControllerToken(),
		),
	}

	log.Println(
		api.NewServer(
			apiConfig,
			baseEndpoints,
			[]api.Endpoints{
				events.NewEndpoints(baseEndpoints, eventsService),
				projects.NewEndpoints(baseEndpoints, projectsService),
				serviceaccounts.NewEndpoints(baseEndpoints, serviceAccountsService),
				sessions.NewEndpoints(baseEndpoints, sessionsService),
				users.NewEndpoints(baseEndpoints, usersService),
			},
		).ListenAndServe(),
	)
}
