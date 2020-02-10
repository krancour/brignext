package main

import (
	"log"

	"github.com/krancour/brignext/pkg/api"
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

	userStore, err := mongodb.NewUserStore(database)
	if err != nil {
		log.Fatal(err)
	}

	projectStore, err := mongodb.NewProjectStore(database)
	if err != nil {
		log.Fatal(err)
	}

	logStore := mongodb.NewLogStore(database)

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
			userStore,
			projectStore,
			logStore,
		).ListenAndServe(),
	)
}
