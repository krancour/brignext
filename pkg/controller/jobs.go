package controller

import (
	"time"

	"github.com/brigadecore/brigade/pkg/brigade"
	oldStorage "github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func jobsIndexInformer(
	client kubernetes.Interface,
	namespace string,
) cache.SharedIndexInformer {
	labelSelector := labels.SelectorFromSet(
		labels.Set{
			"heritage":  "brigade",
			"component": "job",
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

func (c *controller) syncNewJob(obj interface{}) {
	c.syncJob(obj)
}

func (c *controller) syncUpdatedJob(_, updatedObj interface{}) {
	c.syncJob(updatedObj)
}

func (c *controller) syncJob(obj interface{}) {
	jobPod := obj.(*corev1.Pod)

	// We're not interested in processing things that are deleted or already
	// in the process of being deleted
	if jobPod.DeletionTimestamp != nil {
		return
	}

	job := oldStorage.NewJobFromPod(*jobPod)

	existingJob, err := c.projectStore.GetJob(job.ID)
	if err != nil {
		glog.Errorf(
			"error searching new store for possible existing job %q: %s",
			job.ID,
			err,
		)
		return
	}

	buildID := jobPod.Labels["build"]
	if buildID == "" {
		glog.Errorf("cannot sync job %q because it has no build id", job.ID)
		return
	}

	if existingJob == nil {
		err = c.projectStore.CreateJob(buildID, job)
	} else {
		err = c.projectStore.UpdateJobStatus(job.ID, string(job.Status))
	}
	if err != nil {
		glog.Errorf("error syncing job %q to new store: %s", job.ID, err)
		return
	}

	glog.Infof("synced job %q to new store", job.ID)

	if job.Status == brigade.JobSucceeded || job.Status == brigade.JobFailed {
		go func() {
			<-time.After(time.Minute)
			if err := c.podsClient.Delete(jobPod.Name, nil); err != nil {
				glog.Errorf("error deleting job %q pod: %s", job.ID, err)
			}
		}()
	}
}
