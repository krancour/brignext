package controller

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/client"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

type Controller interface {
	Run(context.Context) error
}

type controller struct {
	controllerConfig Config
	apiClient        client.Client
	redisClient      *redis.Client
	kubeClient       *kubernetes.Clientset

	// All of the following behaviors can be overridden for testing purposes
	manageWorkers                     func(context.Context)
	resumeWorkers                     func(context.Context) error
	manageProjectWorkerQueueConsumers func(context.Context)

	workerContextCh chan workerContext
	// All goroutines we launch will send errors here
	errCh chan error

	wg *sync.WaitGroup
}

func NewController(
	controllerConfig Config,
	apiClient client.Client,
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Controller {
	c := &controller{
		controllerConfig: controllerConfig,
		apiClient:        apiClient,
		redisClient:      redisClient,
		kubeClient:       kubeClient,
		workerContextCh:  make(chan workerContext),
		errCh:            make(chan error),
		wg:               &sync.WaitGroup{},
	}

	// Behaviors
	c.manageWorkers = c.defaultManageWorkers
	c.resumeWorkers = c.defaultResumeWorkers
	c.manageProjectWorkerQueueConsumers =
		c.defaultManageProjectWorkerQueueConsumers

	return c
}

func (c *controller) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Run the main control loop
	c.wg.Add(1)
	go c.manageWorkers(ctx)

	// Synchronously resume monitoring any in-progress workers. This will block
	// until we have capacity to start new workers.
	if err := c.resumeWorkers(ctx); err != nil {
		return errors.Wrap(err, "error resuming in-progress workers")
	}

	// Monitor for new/deleted projects at a regular interval. Launch or stop
	// new project-specific queue consumers as needed.
	//
	// This is deliberately started last so that there's never an attempt to
	// start new workers until all in-progress workers have been prioritized.
	c.wg.Add(1)
	go c.manageProjectWorkerQueueConsumers(ctx)

	// Wait for an error or a completed context
	var err error
	select {
	case err = <-c.errCh:
		cancel() // Shut it all down
	case <-ctx.Done():
	}

	// Wait for everything to finish
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		c.wg.Wait()
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second): // TODO: Does it matter if this is harcoded?
	}

	return err
}
