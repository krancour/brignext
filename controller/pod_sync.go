package main

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// syncDeletedPod only fires when a pod deletion is COMPLETE. i.e. The pod is
// completely gone.
func (c *controller) syncDeletedPod(obj interface{}) {
	c.podsLock.Lock()
	defer c.podsLock.Unlock()
	pod := obj.(*corev1.Pod)
	// Remove this pod from the set of pods we were tracking for deletion.
	// Managing this set is essential to not leaking memory.
	delete(c.deletingPodsSet, namespacedPodName(pod.Namespace, pod.Name))
}

// deletePod deletes a pod after a 60 second delay. The delay is to ensure any
// log aggregators have a chance to get all logs from a completed pod before it
// is torpedoed.
func (c *controller) deletePod(_ corev1.Pod) {
	<-time.After(60 * time.Second)
	// Can't use the podsClient that is stored as a controller attribute. We
	// need to grab a namespaced one.
	//
	// TODO: Uncomment this. This is just to help me hack without things getting
	// deleted from underneath my feet.
	//
	// podsClient := c.kubeClient.CoreV1().Pods(pod.Namespace)
	// namespacedPodName := namespacedPodName(pod.Namespace, pod.Name)
	// log.Printf("finally deleting pod %s", namespacedPodName)
	// if err :=
	// 	podsClient.Delete(pod.Name, &metav1.DeleteOptions{}); err != nil {
	// 	log.Printf("error deleting pod %s: %s", namespacedPodName, err)
	// }
	//
	// TODO: Also need to delete workspace (PVC)
	// TODO: Also need to delete worker and job configmaps and secrets
	// TODO: When do the event configmap and event secret get deleted???
	// TODO: Maybe we should actually let the API handle all of that! When the
	// worker or job status is updated, delete whatever isn't needed anymore.
}
