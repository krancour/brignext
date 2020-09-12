package main

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// syncDeletedPod only fires when a pod deletion is COMPLETE. i.e. The pod is
// completely gone.
func (o *observer) syncDeletedPod(obj interface{}) {
	o.syncMu.Lock()
	defer o.syncMu.Unlock()
	pod := obj.(*corev1.Pod)
	// Remove this pod from the set of pods we were tracking for deletion.
	// Managing this set is essential to not leaking memory.
	delete(o.deletingPodsSet, namespacedPodName(pod.Namespace, pod.Name))
}

// deletePod deletes a pod after a 60 second delay. The delay is to ensure any
// log aggregators have a chance to get all logs from a completed pod before it
// is torpedoed.
func (o *observer) deletePod(_ corev1.Pod) {
	<-time.After(60 * time.Second)
	// TODO: Also delete any others k8s resources used by this pod.
}
