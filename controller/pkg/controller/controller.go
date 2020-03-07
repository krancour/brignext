package controller

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/client"
	"k8s.io/client-go/kubernetes"
)

type Controller interface {
	Run(context.Context) error
}

type controller struct {
	apiClient   client.Client
	redisClient *redis.Client
	kubeClient  *kubernetes.Clientset
}

func NewController(
	apiClient client.Client,
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Controller {
	return &controller{
		apiClient:   apiClient,
		redisClient: redisClient,
		kubeClient:  kubeClient,
	}
}

func (c *controller) Run(ctx context.Context) error {
	// TODO: Get all projects and start a message consumer for each
	// c.asyncEngine.RegisterJob("executeWorker", c.workerExecute)
	// c.asyncEngine.RegisterJob("monitorWorker", c.workerMonitor)
	select {}
}
