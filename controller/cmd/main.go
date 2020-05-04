package main

import (
	"log"

	"github.com/krancour/brignext/v2/client"
	"github.com/krancour/brignext/v2/controller/pkg/controller"
	"github.com/krancour/brignext/v2/pkg/kubernetes"
	"github.com/krancour/brignext/v2/pkg/redis"
	"github.com/krancour/brignext/v2/pkg/signals"
	"github.com/krancour/brignext/v2/pkg/version"
)

func main() {
	log.Printf(
		"Starting BrigNext Controller -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	config, err := controller.GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	apiClient := client.NewClient(
		config.APIAddress,
		config.APIToken,
		config.IgnoreAPICertWarnings,
	)

	redisClient, err := redis.Client()
	if err != nil {
		log.Fatal(err)
	}

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	controller := controller.NewController(
		config,
		apiClient,
		redisClient,
		kubeClient,
	)

	log.Println(controller.Run(signals.Context()))
}
