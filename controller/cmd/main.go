package main

import (
	"log"

	async "github.com/deis/async/redis"
	"github.com/krancour/brignext/client"
	"github.com/krancour/brignext/controller/pkg/controller"
	"github.com/krancour/brignext/pkg/kubernetes"
	"github.com/krancour/brignext/pkg/signals"
	"github.com/krancour/brignext/pkg/version"
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

	asyncConfig, err := async.GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	asyncEngine := async.NewEngine(asyncConfig)

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	controller := controller.NewController(apiClient, asyncEngine, kubeClient)

	log.Println(controller.Run(signals.Context()))
}
