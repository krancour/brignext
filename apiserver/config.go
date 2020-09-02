package main

// nolint: lll
import (
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/serviceaccounts"
	serviceaccountsMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/serviceaccounts/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/sessions"
	sessionsMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/sessions/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/users"
	usersMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/users/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/core/events"
	eventsKubernetes "github.com/krancour/brignext/v2/apiserver/internal/core/events/kubernetes"
	eventsMongodb "github.com/krancour/brignext/v2/apiserver/internal/core/events/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/core/projects"
	projectsMongodb "github.com/krancour/brignext/v2/apiserver/internal/core/projects/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/queue/amqp"
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

	// Projects
	projectsStore, err := projectsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	projectsService := projects.NewService(
		projectsStore,
		usersStore,
		serviceAccountsStore,
		projects.NewScheduler(kubeClient),
	)

	// Events-- depends on projects
	queueWriterFactory, err := amqp.GetQueueWriterFactoryFromEnvironment()
	if err != nil {
		return nil, err
	}
	eventsStore, err := eventsMongodb.NewStore(database)
	if err != nil {
		return nil, err
	}
	schedulerConfig, err := events.GetConfigFromEnvironment()
	if err != nil {
		return nil, err
	}
	scheduler := events.NewScheduler(
		schedulerConfig,
		queueWriterFactory,
		kubeClient,
	)
	eventsService := events.NewService(
		projectsStore,
		eventsStore,
		eventsKubernetes.NewLogsStore(kubeClient),
		eventsMongodb.NewLogsStore(database),
		scheduler,
	)

	baseEndpoints := &apimachinery.BaseEndpoints{
		TokenAuthFilter: authn.NewTokenAuthFilter(
			sessionsService.GetByToken,
			eventsService.GetByWorkerToken,
			usersService.Get,
			serviceAccountsService.GetByToken,
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
