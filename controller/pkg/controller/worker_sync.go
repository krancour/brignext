package controller

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (c *controller) defaultSyncExistingWorkerPods() error {
	workerPodList, err := c.podsClient.List(
		metav1.ListOptions{
			LabelSelector: c.workerPodsSelector.String(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "error listing pods")
	}
	for _, workerPod := range workerPodList.Items {
		c.syncWorkerPod(&workerPod)
	}
	return nil
}

func (c *controller) defaultContinuouslySyncWorkerPods(ctx context.Context) {
	workerPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = c.workerPodsSelector.String()
				return c.podsClient.List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = c.workerPodsSelector.String()
				return c.podsClient.Watch(options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	workerPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.syncWorkerPod,
			UpdateFunc: func(_, newObj interface{}) {
				c.syncWorkerPod(newObj)
			},
			DeleteFunc: c.syncDeletedPod,
		},
	)
	workerPodsInformer.Run(ctx.Done())
}

func (c *controller) syncWorkerPod(obj interface{}) {
	c.podsLock.Lock()
	defer c.podsLock.Unlock()
	workerPod := obj.(*corev1.Pod)
	namespacedWorkerPodName := namespacedPodName(
		workerPod.Namespace,
		workerPod.Name,
	)
	if workerPod.Status.Phase == corev1.PodSucceeded ||
		workerPod.Status.Phase == corev1.PodFailed ||
		workerPod.DeletionTimestamp != nil {
		// If the worker pod has run to completion or has been deleted, stop
		// counting it as one that is taking up available capacity.
		delete(c.workerPodsSet, namespacedWorkerPodName)

		// Use the API to update worker status so it corresponds to worker pod
		// status
		if workerPod.Status.Phase == corev1.PodSucceeded ||
			workerPod.Status.Phase == corev1.PodFailed {
			eventID := workerPod.Labels["brignext.io/event"]
			workerName := workerPod.Labels["brignext.io/worker"]
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var status brignext.WorkerStatus
			if workerPod.Status.Phase == corev1.PodSucceeded {
				status = brignext.WorkerStatusSucceeded
			} else {
				status = brignext.WorkerStatusFailed
			}
			if err := c.apiClient.UpdateEventWorkerStatus(
				ctx,
				eventID,
				workerName,
				status,
			); err != nil {
				// TODO: Can we return this over the errCh somehow? Only problem is we
				// don't want to block forever and we don't have access to the context
				// here. Maybe we can make the context an attribute of the controller?
				log.Printf(
					"error updating status for event %q worker %q: %s",
					eventID,
					workerName,
					err,
				)
			}
		}

		if workerPod.DeletionTimestamp == nil {
			// We want to delete this pod after a short delay, but first let's make
			// sure we aren't already working on that. If we schedule this for
			// deletion more than once, we'll end up causing some errors.
			_, alreadyDeleting := c.deletingPodsSet[namespacedWorkerPodName]
			if !alreadyDeleting {
				log.Printf("scheduling pod %s deletion\n", namespacedWorkerPodName)
				c.deletingPodsSet[namespacedWorkerPodName] = struct{}{}
				// Do NOT pass the pointer. It seems to be reused by the informer.
				// Pass the thing it points TO.
				go c.deletePod(*workerPod)
			}
		}
	} else {
		// Make sure this worker pod is counted as one that is taking up available
		// capacity.
		c.workerPodsSet[namespacedWorkerPodName] = struct{}{}
	}
}
