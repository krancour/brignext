package main

import (
	"log"

	async "github.com/deis/async/redis"
	"github.com/golang/glog"
	"github.com/krancour/brignext/pkg/api"
	"github.com/krancour/brignext/pkg/kubernetes"
	mongodbUtils "github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/oidc"
	"github.com/krancour/brignext/pkg/scheduler"
	"github.com/krancour/brignext/pkg/service"
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
	asyncConfig, err := async.GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	asyncEngine := async.NewEngine(asyncConfig)
	kubeClient, err := kubernetes.Client()
	if err != nil {
		glog.Fatal(err)
	}
	brignextNamespace, err := kubernetes.BrigNextNamespace()
	if err != nil {
		glog.Fatal(err)
	}
	scheduler := scheduler.NewScheduler(
		asyncEngine,
		kubeClient,
		brignextNamespace,
	)

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
