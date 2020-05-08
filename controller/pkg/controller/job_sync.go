package controller

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2"
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
				return c.podsClient.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = c.jobPodsSelector.String()
				return c.podsClient.Watch(ctx, options)
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

	// Job pods are only deleted by the API (during event or worker abortion or
	// cancelation) or by the controller. In EITHER case, the job pod has already
	// reached a terminal state and that has already been recorded in the
	// database, so there's nothing to do.
	if jobPod.DeletionTimestamp != nil {
		return
	}

	// Use the API to update job phase so it corresponds to job pod phase
	eventID := jobPod.Labels["brignext.io/event"]
	jobName := jobPod.Labels["brignext.io/job"]

	status := brignext.JobStatus{}
	switch jobPod.Status.Phase {
	case corev1.PodPending:
		// This pod is on its way up. For Brigade's purposes, this counts as
		// running.
		status.Phase = brignext.JobPhaseRunning
	case corev1.PodRunning:
		status.Phase = brignext.JobPhaseRunning
	case corev1.PodSucceeded:
		status.Phase = brignext.JobPhaseSucceeded
	case corev1.PodFailed:
		status.Phase = brignext.JobPhaseFailed
	case corev1.PodUnknown:
		status.Phase = brignext.JobPhaseUnknown
	}

	if jobPod.Status.StartTime != nil {
		status.Started = &jobPod.Status.StartTime.Time
	}
	if len(jobPod.Status.ContainerStatuses) > 0 &&
		jobPod.Status.ContainerStatuses[0].State.Terminated != nil {
		status.Ended =
			&jobPod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.apiClient.UpdateJobStatus(
		ctx,
		eventID,
		jobName,
		status,
	); err != nil {
		// TODO: Can we return this over the errCh somehow? Only problem is we
		// don't want to block forever and we don't have access to the context
		// here. Maybe we can make the context an attribute of the controller?
		log.Printf(
			"error updating status for event %q worker job %q: %s",
			eventID,
			jobName,
			err,
		)
	}

	if jobPod.Status.Phase == corev1.PodSucceeded ||
		jobPod.Status.Phase == corev1.PodFailed {
		namespacedJobPodName := namespacedPodName(jobPod.Namespace, jobPod.Name)
		// We want to delete this pod after a short delay, but first let's make
		// sure we aren't already working on that. If we schedule this for
		// deletion more than once, we'll end up causing some errors.
		_, alreadyDeleting := c.deletingPodsSet[namespacedJobPodName]
		if !alreadyDeleting {
			log.Printf("scheduling job pod %s deletion\n", namespacedJobPodName)
			c.deletingPodsSet[namespacedJobPodName] = struct{}{}
			// Do NOT pass the pointer. It seems to be reused by the informer.
			// Pass the thing it points TO.
			go c.deletePod(*jobPod)
		}
	}
}
