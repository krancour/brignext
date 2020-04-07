package controller

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (c *controller) defaultContinuouslySyncJobPods(ctx context.Context) {
	jobPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = c.jobPodsSelector.String()
				return c.podsClient.List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = c.jobPodsSelector.String()
				return c.podsClient.Watch(options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	jobPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.syncJobPod,
			UpdateFunc: func(_, newObj interface{}) {
				c.syncJobPod(newObj)
			},
			DeleteFunc: c.syncDeletedPod,
		},
	)
	jobPodsInformer.Run(ctx.Done())
}

func (c *controller) syncJobPod(obj interface{}) {
	c.podsLock.Lock()
	defer c.podsLock.Unlock()
	jobPod := obj.(*corev1.Pod)
	namespacedJobPodName := namespacedPodName(jobPod.Namespace, jobPod.Name)
	if jobPod.Status.Phase == corev1.PodSucceeded ||
		jobPod.Status.Phase == corev1.PodFailed ||
		jobPod.DeletionTimestamp != nil {
		// Use the API to update job status so it corresponds to job pod status
		if jobPod.Status.Phase == corev1.PodSucceeded ||
			jobPod.Status.Phase == corev1.PodFailed {
			eventID := jobPod.Labels["brignext.io/event"]
			workerName := jobPod.Labels["brignext.io/worker"]
			jobName := jobPod.Labels["brignext.io/job"]
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var status brignext.JobStatus
			if jobPod.Status.Phase == corev1.PodSucceeded {
				status = brignext.JobStatusSucceeded
			} else {
				status = brignext.JobStatusFailed
			}
			if err := c.apiClient.UpdateJobStatus(
				ctx,
				eventID,
				workerName,
				jobName,
				status,
			); err != nil {
				// TODO: Can we return this over the errCh somehow? Only problem is we
				// don't want to block forever and we don't have access to the context
				// here. Maybe we can make the context an attribute of the controller?
				log.Printf(
					"error updating status for event %q worker %q job %q: %s",
					eventID,
					workerName,
					jobName,
					err,
				)
			}
		}

		if jobPod.DeletionTimestamp == nil {
			// We want to delete this pod after a short delay, but first let's make
			// sure we aren't already working on that. If we schedule this for
			// deletion more than once, we'll end up causing some errors.
			_, alreadyDeleting := c.deletingPodsSet[namespacedJobPodName]
			if !alreadyDeleting {
				log.Printf("scheduling pod %s deletion\n", namespacedJobPodName)
				c.deletingPodsSet[namespacedJobPodName] = struct{}{}
				// Do NOT pass the pointer. It seems to be reused by the informer.
				// Pass the thing it points TO.
				go c.deletePod(*jobPod)
			}
		}
	}
}
