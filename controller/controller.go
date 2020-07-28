package main

import (
	"context"
	"sync"
	"time"

	"github.com/krancour/brignext/v2/internal/events"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	workerPodsSelector = labels.Set(
		map[string]string{
			"brignext.io/component": "worker",
		},
	).AsSelector().String()

	jobPodsSelector = labels.Set(
		map[string]string{
			"brignext.io/component": "job",
		},
	).AsSelector().String()
)

type Controller interface {
	Run(context.Context) error
}

// TODO: Bust this up into two separate components-- scheduler and observer
type controller struct {
	controllerConfig Config
	apiClient        api.Client
	// TODO: This should be closed somewhere
	eventReceiverFactory events.ReceiverFactory
	kubeClient           *kubernetes.Clientset
	podsClient           corev1.PodInterface
	workerPodsSet        map[string]struct{}
	deletingPodsSet      map[string]struct{}
	podsLock             sync.Mutex
	availabilityCh       chan struct{}
	errCh                chan error // All goroutines will send fatal errors here
}

func NewController(
	controllerConfig Config,
	apiClient api.Client,
	eventReceiverFactory events.ReceiverFactory,
	kubeClient *kubernetes.Clientset,
) Controller {
	podsClient := kubeClient.CoreV1().Pods("")
	return &controller{
		controllerConfig:     controllerConfig,
		apiClient:            apiClient,
		eventReceiverFactory: eventReceiverFactory,
		kubeClient:           kubeClient,
		podsClient:           podsClient,
		workerPodsSet:        map[string]struct{}{},
		deletingPodsSet:      map[string]struct{}{},
		availabilityCh:       make(chan struct{}),
		errCh:                make(chan error),
	}
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
	// new project-specific event loops as needed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.manageProjectEventLoops(ctx)
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
