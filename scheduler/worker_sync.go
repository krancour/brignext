package main

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var workerPodsSelector = labels.Set(
	map[string]string{
		"brignext.io/component": "worker",
	},
).AsSelector().String()

func (s *scheduler) syncExistingWorkerPods(ctx context.Context) error {
	workerPodList, err := s.podsClient.List(
		ctx,
		metav1.ListOptions{
			LabelSelector: workerPodsSelector,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error listing pods")
	}
	for _, workerPod := range workerPodList.Items {
		s.syncWorkerPod(&workerPod)
	}
	return nil
}

func (s *scheduler) continuouslySyncWorkerPods(ctx context.Context) {
	workerPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = workerPodsSelector
				return s.podsClient.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = workerPodsSelector
				return s.podsClient.Watch(ctx, options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	workerPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: s.syncWorkerPod,
			UpdateFunc: func(_, newObj interface{}) {
				s.syncWorkerPod(newObj)
			},
		},
	)
	workerPodsInformer.Run(ctx.Done())
}

func (s *scheduler) syncWorkerPod(obj interface{}) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	workerPod := obj.(*corev1.Pod)

	namespacedWorkerPodName :=
		namespacedPodName(workerPod.Namespace, workerPod.Name)

	if workerPod.DeletionTimestamp != nil {
		// Make sure this pod isn't counted as consuming capacity
		delete(s.workerPodsSet, namespacedWorkerPodName)
		return
	}

	switch workerPod.Status.Phase {
	case corev1.PodPending:
		// A pending pod is on its way up. We need to count this as consuming
		// capacity
		s.workerPodsSet[namespacedWorkerPodName] = struct{}{}
	case corev1.PodRunning:
		// Make sure this pod IS counted as consuming capacity
		s.workerPodsSet[namespacedWorkerPodName] = struct{}{}
	case corev1.PodSucceeded:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(s.workerPodsSet, namespacedWorkerPodName)
	case corev1.PodFailed:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(s.workerPodsSet, namespacedWorkerPodName)
	case corev1.PodUnknown:
		// Make sure this pod IS counted as consuming capacity... because we just
		// don't know. (If someone or something deletes it, it will all work itself
		// out.)
		s.workerPodsSet[namespacedWorkerPodName] = struct{}{}
	}

}
