package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext/pkg/crypto"
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
	CreateProjectNamespace(pojectID string) (string, error)
	ScheduleEvent(
		eventID string,
		delay time.Duration,
	) error
	// ScheduleWorkers(
	// 	namespace string,
	// 	eventID string,
	// 	workerNames []string,
	// 	delay time.Duration,
	// ) error
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

func (s *sched) ScheduleEvent(
	eventID string,
	delay time.Duration,
) error {
	if err := s.asyncEngine.SubmitTask(
		async.NewDelayedTask(
			"processEvent",
			map[string]string{
				"eventID": eventID,
			},
			delay,
		),
	); err != nil {
		return errors.Wrapf(
			err,
			"error submitting processing task for event %q",
			eventID,
		)
	}
	return nil
}

// func (s *sched) ScheduleWorkers(
// 	namespace string,
// 	eventID string,
// 	workerNames []string,
// 	delay time.Duration,
// ) error {
// 	for _, workerName := range workerNames {
// 		if err := s.asyncEngine.SubmitTask(
// 			async.NewDelayedTask(
// 				"executeWorker",
// 				map[string]string{
// 					"eventID":    eventID,
// 					"workerName": workerName,
// 				},
// 				delay,
// 			),
// 		); err != nil {
// 			return errors.Wrapf(
// 				err,
// 				"error submitting execute task for event %q worker %q",
// 				eventID,
// 				workerName,
// 			)
// 		}
// 	}
// 	return nil
// }

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
