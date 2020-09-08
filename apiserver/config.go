package main

// nolint: lll
import (
	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	authxMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/authx/mongodb"
	authxREST "github.com/brigadecore/brigade/v2/apiserver/internal/authx/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	coreKubernetes "github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/core/mongodb"
	coreREST "github.com/brigadecore/brigade/v2/apiserver/internal/core/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/oidc"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/system"
	systemREST "github.com/brigadecore/brigade/v2/apiserver/internal/system/rest"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
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
	serviceAccountsService :=
		authx.NewServiceAccountsService(serviceAccountsStore)

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

	rolesStore, err := authxMongodb.NewRolesStore(database)
	if err != nil {
		return nil, err
	}

	substrateConfig, err := core.GetConfigFromEnvironment()
	if err != nil {
		return nil, err
	}
	queueWriterFactory, err := amqp.GetQueueWriterFactoryFromEnvironment()
	if err != nil {
		return nil, err
	}

	// Projects
	projectsStore, err := coreMongodb.NewProjectsStore(database)
	if err != nil {
		return nil, err
	}
	secretsStore := coreKubernetes.NewSecretsStore(kubeClient)
	substrate :=
		coreKubernetes.NewSubstrate(substrateConfig, queueWriterFactory, kubeClient)
	projectsService := core.NewProjectsService(
		projectsStore,
		usersStore,
		serviceAccountsStore,
		rolesStore,
		substrate,
	)
	secretsService := core.NewSecretsService(projectsStore, secretsStore)
	projectRolesService := core.NewProjectRolesService(
		projectsStore,
		usersStore,
		serviceAccountsStore,
		rolesStore,
	)

	// Events-- depends on projects
	eventsStore, err := coreMongodb.NewEventsStore(database)
	if err != nil {
		return nil, err
	}
	workersStore, err := coreMongodb.NewWorkersStore(database)
	if err != nil {
		return nil, err
	}
	jobsStore, err := coreMongodb.NewJobsStore(database)
	if err != nil {
		return nil, err
	}
	eventsService := core.NewEventsService(projectsStore, eventsStore, substrate)
	workersService := core.NewWorkersService(eventsStore, workersStore, substrate)
	jobsService := core.NewJobsService(eventsStore, jobsStore, substrate)
	logsService := core.NewLogsService(
		eventsStore,
		coreKubernetes.NewLogsStore(kubeClient),
		coreMongodb.NewLogsStore(database),
	)

	systemRolesService := system.NewRolesService(
		usersStore,
		serviceAccountsStore,
		rolesStore,
	)

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
			authxREST.NewServiceAccountEndpoints(
				baseEndpoints,
				serviceAccountsService,
			),
			authxREST.NewSessionsEndpoints(baseEndpoints, sessionsService),
			authxREST.NewUsersEndpoints(baseEndpoints, usersService),
			coreREST.NewEventsEndpoints(baseEndpoints, eventsService),
			coreREST.NewWorkersEndpoints(baseEndpoints, workersService),
			coreREST.NewJobsEndpoints(baseEndpoints, jobsService),
			coreREST.NewLogsEndpoints(baseEndpoints, logsService),
			coreREST.NewProjectsEndpoints(baseEndpoints, projectsService),
			coreREST.NewSecretsEndpoints(baseEndpoints, secretsService),
			coreREST.NewProjectsRolesEndpoints(baseEndpoints, projectRolesService),
			systemREST.NewRolesEndpoints(baseEndpoints, systemRolesService),
		},
	), nil
}
