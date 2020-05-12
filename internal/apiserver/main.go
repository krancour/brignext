package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/apiserver/api"
	mongodbUtils "github.com/krancour/brignext/v2/internal/apiserver/mongodb"
	"github.com/krancour/brignext/v2/internal/apiserver/oidc"
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
	"github.com/krancour/brignext/v2/internal/apiserver/storage/mongodb"
	"github.com/krancour/brignext/v2/internal/common/kubernetes"
	"github.com/krancour/brignext/v2/internal/common/redis"
	"github.com/krancour/brignext/v2/internal/common/version"
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

	// OpenID Connect config
	oidcConfig, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	// Datastores (Mongo)
	database, err := mongodbUtils.Database()
	if err != nil {
		log.Fatal(err)
	}
	store, err := mongodb.NewStore(database)
	if err != nil {
		log.Fatal(err)
	}
	logStore := mongodb.NewLogStore(database)

	// Scheduler
	redisClient, err := redis.Client()
	if err != nil {
		log.Fatal(err)
	}
	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}
	scheduler := scheduler.NewScheduler(redisClient, kubeClient)

	service := service.NewService(store, scheduler, logStore)

	log.Println(
		api.NewServer(
			apiConfig,
			oidcConfig,
			oidcIdentityVerifier,
			service,
		).ListenAndServe(),
	)
}
