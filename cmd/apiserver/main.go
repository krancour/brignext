package main

import (
	"log"

	"github.com/krancour/brignext/pkg/api"
	"github.com/krancour/brignext/pkg/brignext/service"
	mongodbUtils "github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/oidc"
	"github.com/krancour/brignext/pkg/storage/mongodb"
	"github.com/krancour/brignext/pkg/version"
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
	oauth2Config, oidcIdentityVerifier, err :=
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

	service := service.NewService(store, logStore)

	// // TODO: Do something with this
	// // Queues (Redis)
	// redisClient, err := redis.Client()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	log.Println(
		api.NewServer(
			apiConfig,
			oauth2Config,
			oidcIdentityVerifier,
			service,
		).ListenAndServe(),
	)
}
