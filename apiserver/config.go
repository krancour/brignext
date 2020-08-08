package main

// nolint: lll
import (
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/auth"
	"github.com/krancour/brignext/v2/apiserver/internal/events"
	"github.com/krancour/brignext/v2/apiserver/internal/events/amqp"
	eventsMongodb "github.com/krancour/brignext/v2/apiserver/internal/events/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/projects"
	projectsMongodb "github.com/krancour/brignext/v2/apiserver/internal/projects/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/serviceaccounts"
	serviceaccountsMongodb "github.com/krancour/brignext/v2/apiserver/internal/serviceaccounts/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/sessions"
	sessionsMongodb "github.com/krancour/brignext/v2/apiserver/internal/sessions/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/users"
	usersMongodb "github.com/krancour/brignext/v2/apiserver/internal/users/mongodb"
	"github.com/krancour/brignext/v2/internal/kubernetes"
)

func getAPIServerFromEnvironment() (apimachinery.Server, error) {

	// API server config
	apiConfig, err := apimachinery.GetConfigFromEnvironment()
	if err != nil {
		return nil, err
	}

	// Common
	database, err := mongodb.Database()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.Client()
	if err != nil {
		return nil, err
	}

	// Projects
	projectsStore, err := projectsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	projectsService := projects.NewService(
		projectsStore,
		projects.NewScheduler(kubeClient),
	)

	// Events-- depends on projects
	eventsSenderFactory, err := amqp.GetEventsSenderFactoryFromEnvironment()
	if err != nil {
		return nil, err
	}
	eventsStore, err := eventsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	eventsService := events.NewService(
		projectsStore,
		eventsStore,
		eventsMongodb.NewLogsStore(database),
		events.NewScheduler(eventsSenderFactory, kubeClient),
	)

	// Service Accounts
	serviceAccountsStore, err := serviceaccountsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	serviceAccountsService := serviceaccounts.NewService(serviceAccountsStore)

	// Users
	usersStore, err := usersMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	usersService := users.NewService(usersStore)

	// Sessions-- depends on users
	oauth2Config, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		return nil, err
	}
	sessionsStore, err := sessionsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	sessionsService := sessions.NewService(
		sessionsStore,
		usersStore,
		apiConfig.RootUserEnabled(),
		apiConfig.HashedRootUserPassword(),
		oauth2Config,
		oidcIdentityVerifier,
	)

	baseEndpoints := &apimachinery.BaseEndpoints{
		TokenAuthFilter: auth.NewTokenAuthFilter(
			sessionsService.GetByToken,
			usersService.Get,
			apiConfig.RootUserEnabled(),
			apiConfig.HashedSchedulerToken(),
			apiConfig.HashedObserverToken(),
		),
	}

	return apimachinery.NewServer(
		apiConfig,
		baseEndpoints,
		[]apimachinery.Endpoints{
			events.NewEndpoints(baseEndpoints, eventsService),
			projects.NewEndpoints(baseEndpoints, projectsService),
			serviceaccounts.NewEndpoints(baseEndpoints, serviceAccountsService),
			sessions.NewEndpoints(baseEndpoints, sessionsService),
			users.NewEndpoints(baseEndpoints, usersService),
		},
	), nil
}
