package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Controller interface {
	Run(context.Context) error
}

type controller struct {
	controllerConfig   Config
	apiClient          api.Client
	redisClient        *redis.Client
	kubeClient         *kubernetes.Clientset
	podsClient         corev1.PodInterface
	workerPodsSelector labels.Selector
	workerPodsSet      map[string]struct{}
	deletingPodsSet    map[string]struct{}
	podsLock           sync.Mutex
	availabilityCh     chan struct{}
	jobPodsSelector    labels.Selector

	// All of the following behaviors can be overridden for testing purposes
	syncExistingWorkerPods            func(context.Context) error
	manageCapacity                    func(context.Context)
	continuouslySyncWorkerPods        func(context.Context)
	manageProjectWorkerQueueConsumers func(context.Context)
	continuouslySyncJobPods           func(context.Context)

	workerContextCh chan workerContext
	// All goroutines we launch will send errors here
	errCh chan error
}

func NewController(
	controllerConfig Config,
	apiClient api.Client,
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Controller {
	podsClient := kubeClient.CoreV1().Pods("")
	workerPodsSelector := labels.Set(
		map[string]string{
			"brignext.io/component": "worker",
		},
	).AsSelector()
	jobPodsSelector := labels.Set(
		map[string]string{
			"brignext.io/component": "job",
		},
	).AsSelector()
	c := &controller{
		controllerConfig: controllerConfig,
		apiClient:        apiClient,
		redisClient:      redisClient,
		kubeClient:       kubeClient,

		// New stuff
		// TODO: Organize this better
		podsClient:         podsClient,
		workerPodsSelector: workerPodsSelector,
		workerPodsSet:      map[string]struct{}{},
		deletingPodsSet:    map[string]struct{}{},
		availabilityCh:     make(chan struct{}),
		jobPodsSelector:    jobPodsSelector,

		workerContextCh: make(chan workerContext),
		errCh:           make(chan error),
	}

	// Behaviors
	c.syncExistingWorkerPods = c.defaultSyncExistingWorkerPods
	c.manageCapacity = c.defaultManageCapacity
	c.continuouslySyncWorkerPods = c.defaultContinuouslySyncWorkerPods
	c.manageProjectWorkerQueueConsumers =
		c.defaultManageProjectWorkerQueueConsumers
	c.continuouslySyncJobPods = c.defaultContinuouslySyncJobPods

	return c
}

func (c *controller) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Synchronously start tracking existing workers pods. This happens
	// synchronously so that the controller is guaranteed a complete picture of
	// what capacity is available before we start taking on new work.
	if err := c.syncExistingWorkerPods(ctx); err != nil {
		return errors.Wrap(err, "error syncing existing worker pods")
	}

	wg := sync.WaitGroup{}

	// Continuously sync worker pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.continuouslySyncWorkerPods(ctx)
	}()

	// Manage available capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.manageCapacity(ctx)
	}()

	// Continuously sync job pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.continuouslySyncJobPods(ctx)
	}()

	// Monitor for new/deleted projects at a regular interval. Launch or stop
	// new project-specific queue consumers as needed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.manageProjectWorkerQueueConsumers(ctx)
	}()

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
		wg.Wait()
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second): // TODO: Does it matter if this is harcoded?
	}

	return err
}
