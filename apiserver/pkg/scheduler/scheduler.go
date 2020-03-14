package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/apiserver/pkg/crypto"
	"github.com/krancour/brignext/pkg/messaging"
	redisMessaging "github.com/krancour/brignext/pkg/messaging/redis"
	"github.com/pkg/errors"
	core_v1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
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
	CreateProjectNamespace(projectID string) (string, error)
	ScheduleWorker(projectID, eventID, workerName string) error
	AbortWorker(namespace, eventID, workerName string) error
	AbortWorkersByEvent(namespace, eventID string) error
	DeleteProjectNamespace(namespace string) error
}

type sched struct {
	redisClient *redis.Client
	kubeClient  *kubernetes.Clientset
}

func NewScheduler(
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &sched{
		redisClient: redisClient,
		kubeClient:  kubeClient,
	}
}

func (s *sched) CreateProjectNamespace(projectID string) (string, error) {
	namespace := strings.ToLower(
		fmt.Sprintf("brignext-%s-%s", projectID, crypto.NewToken(10)),
	)
	if _, err := s.kubeClient.CoreV1().Namespaces().Create(
		&core_v1.Namespace{
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

	if _, err := s.kubeClient.RbacV1().Roles(namespace).Create(
		&rbac_v1.Role{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "workers",
				Namespace: namespace,
			},
			// TODO: Add the correct rules here
			Rules: []rbac_v1.PolicyRule{},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating role \"workers\" in namespace %q",
			namespace,
		)
	}

	if _, err := s.kubeClient.CoreV1().ServiceAccounts(namespace).Create(
		&core_v1.ServiceAccount{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "workers",
				Namespace: namespace,
			},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating service account \"workers\" in namespace %q",
			namespace,
		)
	}

	if _, err := s.kubeClient.RbacV1().RoleBindings(namespace).Create(
		&rbac_v1.RoleBinding{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "workers",
				Namespace: namespace,
			},
			Subjects: []rbac_v1.Subject{
				rbac_v1.Subject{
					Kind:      "ServiceAccount",
					Name:      "workers",
					Namespace: namespace,
				},
			},
			RoleRef: rbac_v1.RoleRef{
				Kind: "Role",
				Name: "workers",
			},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			namespace,
		)
	}

	if _, err := s.kubeClient.RbacV1().Roles(namespace).Create(
		&rbac_v1.Role{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "jobs",
				Namespace: namespace,
			},
			// TODO: Add the correct rules here
			Rules: []rbac_v1.PolicyRule{},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating role \"jobs\" in namespace %q",
			namespace,
		)
	}

	if _, err := s.kubeClient.CoreV1().ServiceAccounts(namespace).Create(
		&core_v1.ServiceAccount{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "jobs",
				Namespace: namespace,
			},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating service account \"jobs\" in namespace %q",
			namespace,
		)
	}

	if _, err := s.kubeClient.RbacV1().RoleBindings(namespace).Create(
		&rbac_v1.RoleBinding{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "jobs",
				Namespace: namespace,
			},
			Subjects: []rbac_v1.Subject{
				rbac_v1.Subject{
					Kind:      "ServiceAccount",
					Name:      "jobs",
					Namespace: namespace,
				},
			},
			RoleRef: rbac_v1.RoleRef{
				Kind: "Role",
				Name: "jobs",
			},
		},
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			namespace,
		)
	}

	return namespace, nil
}

func (s *sched) ScheduleWorker(projectID, eventID, workerName string) error {
	messageBody, err := json.Marshal(struct {
		Event  string `json:"event"`
		Worker string `json:"worker"`
	}{
		Event:  eventID,
		Worker: workerName,
	})
	if err != nil {
		return errors.Wrapf(
			err,
			"error encoding execution task for event %q worker %q",
			eventID,
			workerName,
		)
	}
	producer := redisMessaging.NewProducer(projectID, s.redisClient, nil)
	// TODO: Fix this
	// There's deliberately a short delay here to minimize the possibility of the
	// controller trying (and failing) to locate this new event before the
	// transaction on the store has become durable.
	if err := producer.Publish(
		messaging.NewDelayedMessage(messageBody, 5*time.Second),
	); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for event %q worker %q",
			eventID,
			workerName,
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
	if err := podsClient.DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	); err != nil {
		return errors.Wrap(err, "error deleting pods")
	}
	return nil
}
