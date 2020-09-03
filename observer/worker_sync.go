package main

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/sdk"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (o *observer) continuouslySyncWorkerPods(ctx context.Context) {
	workerPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = workerPodsSelector
				return o.podsClient.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = workerPodsSelector
				return o.podsClient.Watch(ctx, options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	workerPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: o.syncWorkerPod,
			UpdateFunc: func(_, newObj interface{}) {
				o.syncWorkerPod(newObj)
			},
			DeleteFunc: o.syncDeletedPod,
		},
	)
	workerPodsInformer.Run(ctx.Done())
}

func (o *observer) syncWorkerPod(obj interface{}) {
	o.syncMu.Lock()
	defer o.syncMu.Unlock()
	workerPod := obj.(*corev1.Pod)

	// Worker pods are only deleted by the API (during event or worker abortion or
	// cancelation) or by the observer. In EITHER case, the worker pod has
	// already reached a a terminal state and that has already been recorded in
	// the database, so there's nothing to do.
	if workerPod.DeletionTimestamp != nil {
		return
	}

	// Use the API to update worker phase so it corresponds to worker pod phase
	eventID := workerPod.Labels["brignext.io/event"]

	status := sdk.WorkerStatus{}
	switch workerPod.Status.Phase {
	case corev1.PodPending:
		// For BrigNext's purposes, this counts as running
		status.Phase = sdk.WorkerPhaseRunning
	case corev1.PodRunning:
		status.Phase = sdk.WorkerPhaseRunning
	case corev1.PodSucceeded:
		status.Phase = sdk.WorkerPhaseSucceeded
	case corev1.PodFailed:
		status.Phase = sdk.WorkerPhaseFailed
	case corev1.PodUnknown:
		status.Phase = sdk.WorkerPhaseUnknown
	}

	if workerPod.Status.StartTime != nil {
		status.Started = &workerPod.Status.StartTime.Time
	}
	if len(workerPod.Status.ContainerStatuses) > 0 &&
		workerPod.Status.ContainerStatuses[0].State.Terminated != nil {
		status.Ended =
			&workerPod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := o.apiClient.Events().Workers().UpdateStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		// TODO: Can we return this over the errCh somehow? Only problem is we
		// don't want to block forever and we don't have access to the context
		// here. Maybe we can make the context an attribute of the observer?
		log.Printf(
			"error updating status for event %q worker: %s",
			eventID,
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
		_, alreadyDeleting := o.deletingPodsSet[namespacedWorkerPodName]
		if !alreadyDeleting {
			o.deletingPodsSet[namespacedWorkerPodName] = struct{}{}
			// Do NOT pass the pointer. It seems to be reused by the informer.
			// Pass the thing it points TO.
			go o.deletePod(*workerPod)
		}
	}

}
