package main

// nolint: lll
import (
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	authxMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	coreKubernetes "github.com/krancour/brignext/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/krancour/brignext/v2/apiserver/internal/core/mongodb"
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
	serviceAccountsStore, err := authxMongodb.NewServiceAccountsStore(database)
	if err != nil {
		return nil, err
	}
	serviceAccountsService := authx.NewServiceAccountsService(serviceAccountsStore)

	// Users
	usersStore, err := authxMongodb.NewUsersStore(database)
	if err != nil {
		return nil, err
	}
	usersService := authx.NewUsersService(usersStore)

	// Sessions-- depends on users
	oauth2Config, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		return nil, err
	}
	sessionsStore, err := authxMongodb.NewSessionsStore(database)
	if err != nil {
		return nil, err
	}
	sessionsService := authx.NewSessionsService(
		sessionsStore,
		usersStore,
		apiConfig.RootUserEnabled(),
		apiConfig.HashedRootUserPassword(),
		oauth2Config,
		oidcIdentityVerifier,
	)

	// Projects
	projectsStore, err := coreMongodb.NewProjectsStore(database)
	if err != nil {
		return nil, err
	}
	projectsService := core.NewProjectsService(
		projectsStore,
		usersStore,
		serviceAccountsStore,
		core.NewScheduler(kubeClient),
	)

	// Events-- depends on projects
	queueWriterFactory, err := amqp.GetQueueWriterFactoryFromEnvironment()
	if err != nil {
		return nil, err
	}
	eventsStore, err := coreMongodb.NewEventsStore(database)
	if err != nil {
		return nil, err
	}
	schedulerConfig, err := core.GetConfigFromEnvironment()
	if err != nil {
		return nil, err
	}
	scheduler := core.NewEventsScheduler(
		schedulerConfig,
		queueWriterFactory,
		kubeClient,
	)
	eventsService := core.NewEventsService(
		projectsStore,
		eventsStore,
		coreKubernetes.NewLogsStore(kubeClient),
		coreMongodb.NewLogsStore(database),
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
			core.NewEventsEndpoints(baseEndpoints, eventsService),
			core.NewProjectsEndpoints(baseEndpoints, projectsService),
			authx.NewServiceAccountEndpoints(baseEndpoints, serviceAccountsService),
			authx.NewSessionsEndpoints(baseEndpoints, sessionsService),
			authx.NewUsersEndpoints(baseEndpoints, usersService),
		},
	), nil
}
