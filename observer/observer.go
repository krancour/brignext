package main

import (
	"context"
	"sync"
	"time"

	"github.com/krancour/brignext/v2/sdk/api"
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

type Observer interface {
	Run(context.Context) error
}

type observer struct {
	observerConfig  Config
	apiClient       api.Client
	kubeClient      *kubernetes.Clientset
	podsClient      corev1.PodInterface
	deletingPodsSet map[string]struct{}
	syncMu          sync.Mutex
	availabilityCh  chan struct{}
	errCh           chan error // All goroutines will send fatal errors here
}

func NewObserver(
	observerConfig Config,
	apiClient api.Client,
	kubeClient *kubernetes.Clientset,
) Observer {
	podsClient := kubeClient.CoreV1().Pods("")
	return &observer{
		observerConfig:  observerConfig,
		apiClient:       apiClient,
		kubeClient:      kubeClient,
		podsClient:      podsClient,
		deletingPodsSet: map[string]struct{}{},
		availabilityCh:  make(chan struct{}),
		errCh:           make(chan error),
	}
}

func (o *observer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	// Continuously sync worker pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.continuouslySyncWorkerPods(ctx)
	}()

	// Continuously sync job pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.continuouslySyncJobPods(ctx)
	}()

	// Wait for an error or a completed context
	var err error
	select {
	case err = <-o.errCh:
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