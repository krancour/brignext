package main

import (
	"context"
	"sync"
	"time"

	"github.com/krancour/brignext/v2/scheduler/internal/events"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Scheduler interface {
	Run(context.Context) error
}

// TODO: Bust this up into two separate components-- scheduler and observer
type scheduler struct {
	schedulerConfig Config
	apiClient       api.Client
	// TODO: This should be closed somewhere
	eventsReceiverFactory events.ReceiverFactory
	kubeClient            *kubernetes.Clientset
	podsClient            corev1.PodInterface
	workerPodsSet         map[string]struct{}
	syncMu                *sync.Mutex
	availabilityCh        chan struct{}
	errCh                 chan error // All goroutines will send fatal errors here
}

func NewScheduler(
	schedulerConfig Config,
	apiClient api.Client,
	eventsReceiverFactory events.ReceiverFactory,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	podsClient := kubeClient.CoreV1().Pods("")
	return &scheduler{
		schedulerConfig:       schedulerConfig,
		apiClient:             apiClient,
		eventsReceiverFactory: eventsReceiverFactory,
		kubeClient:            kubeClient,
		podsClient:            podsClient,
		workerPodsSet:         map[string]struct{}{},
		syncMu:                &sync.Mutex{},
		availabilityCh:        make(chan struct{}),
		errCh:                 make(chan error),
	}
}

func (s *scheduler) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Synchronously start tracking existing workers pods. This happens
	// synchronously so that the scheduler is guaranteed a complete picture of
	// what capacity is available before we start taking on new work.
	if err := s.syncExistingWorkerPods(ctx); err != nil {
		return errors.Wrap(err, "error syncing existing worker pods")
	}

	wg := sync.WaitGroup{}

	// Continuously sync worker pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.continuouslySyncWorkerPods(ctx)
	}()

	// Manage available capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageCapacity(ctx)
	}()

	// Monitor for new/deleted projects at a regular interval. Launch or stop
	// new project-specific event loops as needed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageProjectEventLoops(ctx)
	}()

	// Wait for an error or a completed context
	var err error
	select {
	case err = <-s.errCh:
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
