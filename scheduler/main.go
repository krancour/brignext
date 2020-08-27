package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/signals"
	"github.com/krancour/brignext/v2/internal/version"
	"github.com/krancour/brignext/v2/scheduler/internal/queue/amqp"
	"github.com/krancour/brignext/v2/sdk/api"
)

func main() {
	log.Printf(
		"Starting BrigNext Scheduler -- version %s -- commit %s",
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

	queueReaderFactory, err := amqp.GetQueueReaderFactoryFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	scheduler := NewScheduler(
		config,
		apiClient,
		queueReaderFactory,
		kubeClient,
	)

	log.Println(scheduler.Run(signals.Context()))
}
