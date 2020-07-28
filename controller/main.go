package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/events/amqp"
	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/signals"
	"github.com/krancour/brignext/v2/internal/version"
	"github.com/krancour/brignext/v2/sdk/api"
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
	apiClient := api.NewClient(
		config.APIAddress,
		config.APIToken,
		config.IgnoreAPICertWarnings,
	)

	eventReceiverFactory, err := amqp.GetReceiverFactoryFromEnvironment()
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
		eventReceiverFactory,
		kubeClient,
	)

	log.Println(controller.Run(signals.Context()))
}
