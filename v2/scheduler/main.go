package main

import (
	"log"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
	"github.com/brigadecore/brigade/v2/scheduler/internal/queue/amqp"
)

func main() {
	log.Printf(
		"Starting Brigade Scheduler -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	config, err := GetConfigFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	apiClient := core.NewAPIClient(
		config.APIAddress,
		config.APIToken,
		&restmachinery.APIClientOptions{
			AllowInsecureConnections: config.IgnoreAPICertWarnings,
		},
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
