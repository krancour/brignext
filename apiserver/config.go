package main

// nolint: lll
import (
	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	authxMongodb "github.com/krancour/brignext/v2/apiserver/internal/authx/mongodb"
	authxREST "github.com/krancour/brignext/v2/apiserver/internal/authx/rest"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	coreKubernetes "github.com/krancour/brignext/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/krancour/brignext/v2/apiserver/internal/core/mongodb"
	coreREST "github.com/krancour/brignext/v2/apiserver/internal/core/rest"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/mongodb"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/oidc"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/queue/amqp"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/system"
	systemREST "github.com/krancour/brignext/v2/apiserver/internal/system/rest"
	"github.com/krancour/brignext/v2/internal/kubernetes"
)

func getAPIServerFromEnvironment() (restmachinery.Server, error) {

	// API server config
	apiConfig, err := restmachinery.GetConfigFromEnvironment()
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

	baseEndpoints := &restmachinery.BaseEndpoints{
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

	return restmachinery.NewServer(
		apiConfig,
		baseEndpoints,
		[]restmachinery.Endpoints{
			authxREST.NewServiceAccountEndpoints(baseEndpoints, serviceAccountsService),
			authxREST.NewSessionsEndpoints(baseEndpoints, sessionsService),
			authxREST.NewUsersEndpoints(baseEndpoints, usersService),
			coreREST.NewEventsEndpoints(baseEndpoints, eventsService),
			coreREST.NewWorkersEndpoints(baseEndpoints, eventsService),
			coreREST.NewJobsEndpoints(baseEndpoints, eventsService),
			coreREST.NewLogsEndpoints(baseEndpoints, eventsService),
			coreREST.NewProjectsEndpoints(baseEndpoints, projectsService),
			coreREST.NewSecretsEndpoints(baseEndpoints, projectsService),
			coreREST.NewProjectsRolesEndpoints(baseEndpoints, projectsService),
			systemREST.NewSystemRolesEndpoints(baseEndpoints, systemService),
		},
	), nil
}
