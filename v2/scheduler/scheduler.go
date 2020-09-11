package main

import (
	"context"
	"sync"
	"time"

	"github.com/brigadecore/brigade/v2/scheduler/internal/queue"
	"github.com/brigadecore/brigade/v2/sdk/core"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Scheduler interface {
	Run(context.Context) error
}

type scheduler struct {
	schedulerConfig Config
	coreClient      core.APIClient
	// TODO: This should be closed somewhere
	queueReaderFactory   queue.ReaderFactory
	kubeClient           *kubernetes.Clientset
	podsClient           corev1.PodInterface
	workerPodsSet        map[string]struct{}
	jobPodsSet           map[string]struct{}
	syncMu               *sync.Mutex
	workerAvailabilityCh chan struct{}
	jobAvailabilityCh    chan struct{}
	errCh                chan error // All goroutines will send fatal errors here
}

func NewScheduler(
	schedulerConfig Config,
	coreClient core.APIClient,
	queueReaderFactory queue.ReaderFactory,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	podsClient := kubeClient.CoreV1().Pods("")
	return &scheduler{
		schedulerConfig:      schedulerConfig,
		coreClient:           coreClient,
		queueReaderFactory:   queueReaderFactory,
		kubeClient:           kubeClient,
		podsClient:           podsClient,
		workerPodsSet:        map[string]struct{}{},
		jobPodsSet:           map[string]struct{}{},
		syncMu:               &sync.Mutex{},
		workerAvailabilityCh: make(chan struct{}),
		jobAvailabilityCh:    make(chan struct{}),
		errCh:                make(chan error),
	}
}

func (s *scheduler) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Synchronously start tracking existing Worker pods. This happens
	// synchronously so that the scheduler is guaranteed a complete picture of
	// what capacity is available before we start taking on new work.
	if err := s.syncExistingWorkerPods(ctx); err != nil {
		return errors.Wrap(err, "error syncing existing worker pods")
	}

	// Synchronously start tracking existing Job pods. This happens
	// synchronously so that the scheduler is guaranteed a complete picture of
	// what capacity is available before we start taking on new work.
	if err := s.syncExistingJobPods(ctx); err != nil {
		return errors.Wrap(err, "error syncing existing job pods")
	}

	wg := sync.WaitGroup{}

	// Continuously sync worker pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.continuouslySyncWorkerPods(ctx)
	}()

	// Continuously sync Job pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.continuouslySyncJobPods(ctx)
	}()

	// Manage available Worker capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageWorkerCapacity(ctx)
	}()

	// Manage available Job capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageJobCapacity(ctx)
	}()

	// Monitor for new/deleted projects at a regular interval. Launch or stop
	// new project-specific Worker and Job loops as needed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageProjectLoops(ctx)
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
