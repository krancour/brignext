package scheduler

import (
	"time"

	"github.com/deis/async"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	projectLabel = "brignext.io/project"
	eventLabel   = "brignext.io/event"
	workerLabel  = "brignext.io/worker"
)

type Scheduler interface {
	ScheduleWorkers(
		eventID string,
		workerNames []string,
		delay time.Duration,
	) error
	AbortWorker(eventID, workerName string) error
	AbortWorkersByEvent(string) error
	AbortWorkersByProject(string) error
}

type sched struct {
	asyncEngine async.Engine
	podsClient  v1.PodInterface
}

func NewScheduler(
	asyncEngine async.Engine,
	kubeClient *kubernetes.Clientset,
	namespace string,
) Scheduler {
	return &sched{
		asyncEngine: asyncEngine,
		podsClient:  kubeClient.CoreV1().Pods(namespace),
	}
}

func (s *sched) ScheduleWorkers(
	eventID string,
	workerNames []string,
	delay time.Duration,
) error {
	for _, workerName := range workerNames {
		if err := s.asyncEngine.SubmitTask(
			async.NewDelayedTask(
				"executeWorker",
				map[string]string{
					"eventID":    eventID,
					"workerName": workerName,
				},
				delay,
			),
		); err != nil {
			return errors.Wrapf(
				err,
				"error submitting execute task for event %q worker %q",
				eventID,
				workerName,
			)
		}
		if err := s.asyncEngine.SubmitTask(
			async.NewDelayedTask(
				"nannyEvent",
				map[string]string{
					"eventID": eventID,
				},
				delay,
			),
		); err != nil {
			return errors.Wrapf(
				err,
				"error submitting nanny task for event %q",
				eventID,
			)
		}
	}
	return nil
}

func (s *sched) AbortWorker(eventID, workerName string) error {
	return s.deletePodsByLabelsMap(
		map[string]string{
			eventLabel:  eventID,
			workerLabel: workerName,
		},
	)
}

func (s *sched) AbortWorkersByEvent(eventID string) error {
	return s.deletePodsByLabelsMap(
		map[string]string{
			eventLabel: eventID,
		},
	)
}

func (s *sched) AbortWorkersByProject(projectID string) error {
	return s.deletePodsByLabelsMap(
		map[string]string{
			projectLabel: projectID,
		},
	)
}

func (s *sched) deletePodsByLabelsMap(labelsMap map[string]string) error {
	selectorStr := labels.SelectorFromSet(labelsMap).String()
	podList, err := s.podsClient.List(
		meta_v1.ListOptions{
			LabelSelector: selectorStr,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error listing pods with labels %s", selectorStr)
	}
	for _, pod := range podList.Items {
		if err :=
			s.podsClient.Delete(pod.Name, &meta_v1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "error deleting pod %q", pod.Name)
		}
	}
	return nil
}
