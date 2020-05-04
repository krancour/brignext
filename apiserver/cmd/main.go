package main

import (
	"log"

	"github.com/krancour/brignext/v2/apiserver/pkg/api"
	mongodbUtils "github.com/krancour/brignext/v2/apiserver/pkg/mongodb"
	"github.com/krancour/brignext/v2/apiserver/pkg/oidc"
	"github.com/krancour/brignext/v2/apiserver/pkg/scheduler"
	"github.com/krancour/brignext/v2/apiserver/pkg/service"
	"github.com/krancour/brignext/v2/apiserver/pkg/storage/mongodb"
	"github.com/krancour/brignext/v2/pkg/kubernetes"
	"github.com/krancour/brignext/v2/pkg/redis"
	"github.com/krancour/brignext/v2/pkg/version"
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
