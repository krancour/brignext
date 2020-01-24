package controller

import (
	"context"

	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/golang/glog"
	"github.com/krancour/brignext/pkg/storage"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type Controller interface {
	Run(ctx context.Context)
}

type controller struct {
	podsClient      corev1.PodInterface
	secretsClient   corev1.SecretInterface
	buildsInformer  cache.SharedIndexInformer
	workersInformer cache.SharedIndexInformer
	jobsInformer    cache.SharedIndexInformer
	oldStore        oldStorage.Store
	projectStore    storage.ProjectStore
}

func NewController(
	kubeClient kubernetes.Interface,
	namespace string,
	oldStore oldStorage.Store,
	projectStore storage.ProjectStore,
) Controller {
	c := &controller{
		podsClient:      kubeClient.CoreV1().Pods(namespace),
		secretsClient:   kubeClient.CoreV1().Secrets(namespace),
		buildsInformer:  buildsIndexInformer(kubeClient, namespace),
		workersInformer: workersIndexInformer(kubeClient, namespace),
		jobsInformer:    jobsIndexInformer(kubeClient, namespace),
		oldStore:        oldStore,
		projectStore:    projectStore,
	}
	c.buildsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.syncNewBuild,
	})
	c.workersInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.syncNewWorker,
		UpdateFunc: c.syncUpdatedWorker,
	})
	c.jobsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.syncNewJob,
		UpdateFunc: c.syncUpdatedJob,
	})
	return c
}

func (c *controller) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		glog.Infof("Controller is shutting down")
	}()
	glog.Infof("Controller is started")
	go func() {
		c.buildsInformer.Run(ctx.Done())
		cancel()
	}()
	go func() {
		c.workersInformer.Run(ctx.Done())
		cancel()
	}()
	go func() {
		c.jobsInformer.Run(ctx.Done())
		cancel()
	}()
	<-ctx.Done()
}
