package controller

import (
	"github.com/golang/glog"

	oldStorage "github.com/brigadecore/brigade/pkg/storage/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func buildsIndexInformer(
	client kubernetes.Interface,
	namespace string,
) cache.SharedIndexInformer {
	const buildSelector = "type=brigade.sh/build"
	secretsClient := client.CoreV1().Secrets(namespace)
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = buildSelector
				return secretsClient.List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = buildSelector
				return secretsClient.Watch(options)
			},
		},
		&corev1.Secret{},
		0,
		cache.Indexers{},
	)
}

func (c *controller) syncNewBuild(obj interface{}) {
	buildSecret := obj.(*corev1.Secret)
	if buildSecret.Labels["synced"] == "synced" {
		return
	}
	build := oldStorage.NewBuildFromSecret(*buildSecret)
	if existingBuild, err := c.projectStore.GetBuild(build.ID); err != nil {
		glog.Errorf(
			"error searching new store for possible existing build %q: %s",
			build.ID,
			err,
		)
		return
	} else if existingBuild != nil {
		// Nothing to do
		return
	}

	if err := c.projectStore.CreateBuild(build); err != nil {
		glog.Errorf("error syncing new build %q to new store: %s", build.ID, err)
		return
	}

	glog.Infof("synced build %q to new store", build.ID)
}
