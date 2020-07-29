package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/krancour/brignext/v2/internal/signals"
	"github.com/krancour/brignext/v2/internal/version"
	"github.com/krancour/brignext/v2/sdk/api"
)

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
	apiClient := api.NewClient(
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
		apiClient,
		kubeClient,
	)

	log.Println(observer.Run(signals.Context()))
}
