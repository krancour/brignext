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

	namespacedWorkerPodName :=
		namespacedPodName(workerPod.Namespace, workerPod.Name)

	// Worker pods are only deleted by the API (during event or worker abortion or
	// cancelation) or by the controller. In EITHER case, the worker pod has
	// already reached a a terminal state and that has already been recorded in
	// the database, so there's nothing to do EXCEPT ensure this controller is not
	// counted among the pods currently consuming available capacity.
	if workerPod.DeletionTimestamp != nil {
		// Make sure this pod isn't counted as consuming capacity
		delete(c.workerPodsSet, namespacedWorkerPodName)
		return
	}

	// Use the API to update worker status so it corresponds to worker pod status
	eventID := workerPod.Labels["brignext.io/event"]
	workerName := workerPod.Labels["brignext.io/worker"]

	var status brignext.WorkerStatus
	switch workerPod.Status.Phase {
	case corev1.PodPending:
		// A pending pod is on its way up. We need to count this as consuming
		// capacity
		c.workerPodsSet[namespacedWorkerPodName] = struct{}{}

		// For Brigade's purposes, this counts as running
		status = brignext.WorkerStatusRunning
	case corev1.PodRunning:
		// Make sure this pod IS counted as consuming capacity
		c.workerPodsSet[namespacedWorkerPodName] = struct{}{}

		status = brignext.WorkerStatusRunning
	case corev1.PodSucceeded:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(c.workerPodsSet, namespacedWorkerPodName)

		status = brignext.WorkerStatusSucceeded
	case corev1.PodFailed:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(c.workerPodsSet, namespacedWorkerPodName)

		status = brignext.WorkerStatusFailed
	case corev1.PodUnknown:
		// Make sure this pod IS counted as consuming capacity... because we just
		// don't know. (If someone or something deletes it, it will all work itself
		// out.)
		c.workerPodsSet[namespacedWorkerPodName] = struct{}{}

		status = brignext.WorkerStatusUnknown
	}

	var started *time.Time
	var ended *time.Time
	if workerPod.Status.StartTime != nil {
		started = &workerPod.Status.StartTime.Time
	}
	if len(workerPod.Status.ContainerStatuses) > 0 &&
		workerPod.Status.ContainerStatuses[0].State.Terminated != nil {
		ended =
			&workerPod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.apiClient.UpdateWorkerStatus(
		ctx,
		eventID,
		workerName,
		started,
		ended,
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

	if workerPod.Status.Phase == corev1.PodSucceeded ||
		workerPod.Status.Phase == corev1.PodFailed {
		namespacedWorkerPodName :=
			namespacedPodName(workerPod.Namespace, workerPod.Name)
		// We want to delete this pod after a short delay, but first let's make
		// sure we aren't already working on that. If we schedule this for
		// deletion more than once, we'll end up causing some errors.
		_, alreadyDeleting := c.deletingPodsSet[namespacedWorkerPodName]
		if !alreadyDeleting {
			log.Printf("scheduling worker pod %s deletion\n", namespacedWorkerPodName)
			c.deletingPodsSet[namespacedWorkerPodName] = struct{}{}
			// Do NOT pass the pointer. It seems to be reused by the informer.
			// Pass the thing it points TO.
			go c.deletePod(*workerPod)
		}
	}

}
