package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext/apiserver/pkg/crypto"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	projectLabel = "brignext.io/project"
	eventLabel   = "brignext.io/event"
	workerLabel  = "brignext.io/worker"
)

type Scheduler interface {
	// TODO: This should move somewhere else-- it's not the scheduler's
	// responsibility
	CreateProjectNamespace(pojectID string) (string, error)
	ScheduleWorker(eventID, workerName string) error
	AbortWorker(namespace, eventID, workerName string) error
	AbortWorkersByEvent(namespace, eventID string) error
	DeleteProjectNamespace(namespace string) error
}

type sched struct {
	asyncEngine async.Engine
	kubeClient  *kubernetes.Clientset
}

func NewScheduler(
	asyncEngine async.Engine,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &sched{
		asyncEngine: asyncEngine,
		kubeClient:  kubeClient,
	}
}

func (s *sched) CreateProjectNamespace(projectID string) (string, error) {
	namespace := strings.ToLower(
		fmt.Sprintf("brignext-%s-%s", projectID, crypto.NewToken(10)),
	)
	if _, err := s.kubeClient.CoreV1().Namespaces().Create(
		&v1.Namespace{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: namespace,
			},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating namespace %q for project %q",
			namespace,
			projectID,
		)
	}
	return namespace, nil
}

func (s *sched) ScheduleWorker(eventID, workerName string) error {
	// TODO: Fix this
	// There's deliberately a short delay here to minimize the possibility of the
	// controller trying (and failing) to locate this new event before the
	// transaction on the store has become durable.
	if err := s.asyncEngine.SubmitTask(
		async.NewDelayedTask(
			"executeWorker",
			map[string]string{
				"event":  eventID,
				"worker": workerName,
			},
			5*time.Second,
		),
	); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for worker %q",
			eventID,
		)
	}
	return nil
}

func (s *sched) AbortWorker(namespace, eventID, workerName string) error {
	return s.deletePodsByLabelsMap(
		namespace,
		map[string]string{
			eventLabel:  eventID,
			workerLabel: workerName,
		},
	)
}

func (s *sched) AbortWorkersByEvent(namespace, eventID string) error {
	return s.deletePodsByLabelsMap(
		namespace,
		map[string]string{
			eventLabel: eventID,
		},
	)
}

func (s *sched) DeleteProjectNamespace(namespace string) error {
	if err := s.kubeClient.CoreV1().Namespaces().Delete(
		namespace,
		&meta_v1.DeleteOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting namespace %q",
			namespace,
		)
	}
	return nil
}

func (s *sched) deletePodsByLabelsMap(
	namespace string,
	labelsMap map[string]string,
) error {
	podsClient := s.kubeClient.CoreV1().Pods(namespace)
	selectorStr := labels.SelectorFromSet(labelsMap).String()
	podList, err := podsClient.List(
		meta_v1.ListOptions{
			LabelSelector: selectorStr,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error listing pods with labels %s", selectorStr)
	}
	for _, pod := range podList.Items {
		if err :=
			podsClient.Delete(pod.Name, &meta_v1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "error deleting pod %q", pod.Name)
		}
	}
	return nil
}
