package controller

import (
	"context"

	"github.com/deis/async"
	"github.com/krancour/brignext/client"
	"k8s.io/client-go/kubernetes"
)

type Controller interface {
	Run(context.Context) error
}

type controller struct {
	apiClient   client.Client
	asyncEngine async.Engine
	kubeClient  *kubernetes.Clientset
}

func NewController(
	apiClient client.Client,
	asyncEngine async.Engine,
	kubeClient *kubernetes.Clientset,
) Controller {
	c := &controller{
		apiClient:   apiClient,
		asyncEngine: asyncEngine,
		kubeClient:  kubeClient,
	}
	c.asyncEngine.RegisterJob("executeWorker", c.workerExecute)
	c.asyncEngine.RegisterJob("monitorWorker", c.workerMonitor)
	return c
}

func (c *controller) Run(ctx context.Context) error {
	// TODO: Also run a healthcheck endpoint
	return c.asyncEngine.Run(ctx)
}
