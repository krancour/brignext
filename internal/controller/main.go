package main

import (
	"log"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/common/kubernetes"
	"github.com/krancour/brignext/v2/internal/common/redis"
	"github.com/krancour/brignext/v2/internal/common/signals"
	"github.com/krancour/brignext/v2/internal/common/version"
)

func main() {
	log.Printf(
		"Starting BrigNext Controller -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	config, err := GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	apiClient := brignext.NewClient(
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

	controller := NewController(
		config,
		apiClient,
		redisClient,
		kubeClient,
	)

	log.Println(controller.Run(signals.Context()))
}
