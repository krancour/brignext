package controller

import (
	"time"

	"github.com/brigadecore/brigade/pkg/brigade"
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	oldKubeStorage "github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func workersIndexInformer(
	client kubernetes.Interface,
	namespace string,
) cache.SharedIndexInformer {
	labelSelector := labels.SelectorFromSet(
		labels.Set{
			"heritage":  "brigade",
			"component": "build",
		},
	)
	podsClient := client.CoreV1().Pods(namespace)
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = labelSelector.String()
				return podsClient.List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = labelSelector.String()
				return podsClient.Watch(options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
}

func (c *controller) syncNewWorker(obj interface{}) {
	c.syncWorker(obj)
}

func (c *controller) syncUpdatedWorker(_, updatedObj interface{}) {
	c.syncWorker(updatedObj)
}

func (c *controller) syncWorker(obj interface{}) {
	workerPod := obj.(*corev1.Pod)

	// We're not interested in processing things that are deleted or already
	// in the process of being deleted
	if workerPod.DeletionTimestamp != nil {
		return
	}

	worker := oldKubeStorage.NewWorkerFromPod(*workerPod)

	if err := c.projectStore.UpdateWorker(worker); err != nil {
		glog.Errorf(
			"error updating build %q worker status in new store: %s",
			worker.BuildID,
			err,
		)
		return
	}

	glog.Infof("synced build %q worker status to new store", worker.BuildID)

	if worker.Status == brigade.JobSucceeded ||
		worker.Status == brigade.JobFailed {
		go func() {
			<-time.After(time.Minute)
			if err := c.podsClient.Delete(workerPod.Name, nil); err != nil {
				glog.Errorf("error deleting worker %q pod: %s", worker.ID, err)
			}
			if err := c.oldStore.DeleteBuild(
				worker.BuildID,
				oldStorage.DeleteBuildOptions{
					SkipRunningBuilds: true,
				},
			); err != nil {
				glog.Errorf("error deleting build secret %q: %s", worker.BuildID, err)
			}
		}()
	}
}
