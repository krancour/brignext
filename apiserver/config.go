package main

// nolint: lll
import (
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	authxAPI "github.com/krancour/brignext/v2/apiserver/internal/authx/api"
	authxMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	coreAPI "github.com/krancour/brignext/v2/apiserver/internal/core/api"
	coreKubernetes "github.com/krancour/brignext/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/krancour/brignext/v2/apiserver/internal/core/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/apimachinery/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/queue/amqp"
	"github.com/krancour/brignext/v2/apiserver/internal/system"
	systemAPI "github.com/krancour/brignext/v2/apiserver/internal/system/api"
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

	systemService := system.NewService(usersStore, serviceAccountsStore)

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
			authxAPI.NewServiceAccountEndpoints(baseEndpoints, serviceAccountsService),
			authxAPI.NewSessionsEndpoints(baseEndpoints, sessionsService),
			authxAPI.NewUsersEndpoints(baseEndpoints, usersService),
			coreAPI.NewEventsEndpoints(baseEndpoints, eventsService),
			coreAPI.NewWorkersEndpoints(baseEndpoints, eventsService),
			coreAPI.NewJobsEndpoints(baseEndpoints, eventsService),
			coreAPI.NewLogsEndpoints(baseEndpoints, eventsService),
			coreAPI.NewProjectsEndpoints(baseEndpoints, projectsService),
			coreAPI.NewSecretsEndpoints(baseEndpoints, projectsService),
			coreAPI.NewProjectsRolesEndpoints(baseEndpoints, projectsService),
			systemAPI.NewSystemRolesEndpoints(baseEndpoints, systemService),
		},
	), nil
}
