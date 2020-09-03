package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/signals"
	"github.com/krancour/brignext/v2/internal/version"
	"github.com/krancour/brignext/v2/sdk/core/api"
)

// TODO: Observer needs functionality for timing out workers and jobs
func main() {
	log.Printf(
		"Starting BrigNext Observer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	config, err := GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	workersClient := api.NewWorkersClient(
		config.APIAddress,
		config.APIToken,
		config.IgnoreAPICertWarnings,
	)

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	observer := NewObserver(
		config,
		workersClient,
		kubeClient,
	)

	log.Println(observer.Run(signals.Context()))
}
