package main

import (
	"context"

	myk8s "github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var jobPodsSelector = labels.Set(
	map[string]string{
		myk8s.LabelComponent: "job",
	},
).AsSelector().String()

func (s *scheduler) syncExistingJobPods(ctx context.Context) error {
	jobPods, err := s.podsClient.List(
		ctx,
		metav1.ListOptions{
			LabelSelector: jobPodsSelector,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error listing job pods")
	}
	for _, jobPod := range jobPods.Items {
		s.syncJobPod(&jobPod)
	}
	return nil
}

func (s *scheduler) continuouslySyncJobPods(ctx context.Context) {
	jobPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = jobPodsSelector
				return s.podsClient.List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = jobPodsSelector
				return s.podsClient.Watch(ctx, options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	jobPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: s.syncJobPod,
			UpdateFunc: func(_, newObj interface{}) {
				s.syncJobPod(newObj)
			},
		},
	)
	jobPodsInformer.Run(ctx.Done())
}

func (s *scheduler) syncJobPod(obj interface{}) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	jobPod := obj.(*corev1.Pod)

	namespacedJobPodName :=
		namespacedPodName(jobPod.Namespace, jobPod.Name)

	if jobPod.DeletionTimestamp != nil {
		// Make sure this pod isn't counted as consuming capacity
		delete(s.jobPodsSet, namespacedJobPodName)
		return
	}

	switch jobPod.Status.Phase {
	case corev1.PodPending:
		// A pending pod is on its way up. We need to count this as consuming
		// capacity
		s.jobPodsSet[namespacedJobPodName] = struct{}{}
	case corev1.PodRunning:
		// Make sure this pod IS counted as consuming capacity
		s.jobPodsSet[namespacedJobPodName] = struct{}{}
	case corev1.PodSucceeded:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(s.jobPodsSet, namespacedJobPodName)
	case corev1.PodFailed:
		// Make sure this pod IS NOT counted as consuming capacity
		delete(s.jobPodsSet, namespacedJobPodName)
	case corev1.PodUnknown:
		// Make sure this pod IS counted as consuming capacity... because we just
		// don't know. (If someone or something deletes it, it will all work itself
		// out.)
		s.jobPodsSet[namespacedJobPodName] = struct{}{}
	}

}
