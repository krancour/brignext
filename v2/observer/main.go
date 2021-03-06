package main

import (
	"log"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
)

// TODO: Observer needs functionality for timing out workers and jobs
func main() {
	log.Printf(
		"Starting Brigade Observer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	config, err := GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	workersClient := core.NewWorkersClient(
		config.APIAddress,
		config.APIToken,
		&restmachinery.APIClientOptions{
			AllowInsecureConnections: config.IgnoreAPICertWarnings,
		},
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
