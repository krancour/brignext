package scheduler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/crypto"
	"github.com/krancour/brignext/pkg/messaging"
	redisMessaging "github.com/krancour/brignext/pkg/messaging/redis"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	core_v1 "k8s.io/api/core/v1"
	rbac_v1 "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

const (
	componentLabel = "brignext.io/component"
	projectLabel   = "brignext.io/project"
	eventLabel     = "brignext.io/event"
	workerLabel    = "brignext.io/worker"
)

type Scheduler interface {
	CreateProject(project brignext.Project) (brignext.Project, error)
	UpdateProject(project brignext.Project) error
	DeleteProject(project brignext.Project) error

	CreateEvent(event brignext.Event) error
	DeleteEvent(event brignext.Event) error

	DeleteWorker(event brignext.Event, workerName string) error
}

type scheduler struct {
	redisClient *redis.Client
	kubeClient  *kubernetes.Clientset
}

func NewScheduler(
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &scheduler{
		redisClient: redisClient,
		kubeClient:  kubeClient,
	}
}

func (s *scheduler) CreateProject(
	project brignext.Project,
) (brignext.Project, error) {
	// Create a unique namespace name for the project
	project.Kubernetes = &brignext.ProjectKubernetesConfig{
		Namespace: strings.ToLower(
			fmt.Sprintf("brignext-%s-%s", project.ID, crypto.NewToken(10)),
		),
	}

	// Create a the project's namespace
	if _, err := s.kubeClient.CoreV1().Namespaces().Create(
		&core_v1.Namespace{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: project.Kubernetes.Namespace,
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating namespace %q for project %q",
			project.Kubernetes.Namespace,
			project.ID,
		)
	}

	// Create an RBAC role for use by all of the project's workers
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		&rbac_v1.Role{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "workers",
			},
			Rules: []rbac_v1.PolicyRule{
				rbac_v1.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"configmaps", "secrets", "pods", "pods/log"},
					Verbs:     []string{"create", "get", "list", "watch"},
				},
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the project's workers
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		&core_v1.ServiceAccount{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "workers",
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating service account \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the workers service account with
	// the workers RBAC role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		&rbac_v1.RoleBinding{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "workers",
			},
			Subjects: []rbac_v1.Subject{
				rbac_v1.Subject{
					Kind:      "ServiceAccount",
					Name:      "workers",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbac_v1.RoleRef{
				Kind: "Role",
				Name: "workers",
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role for use by all of the project's jobs
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		&rbac_v1.Role{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "jobs",
			},
			// TODO: Add the correct rules here
			Rules: []rbac_v1.PolicyRule{},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the project's workers
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		&core_v1.ServiceAccount{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "jobs",
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating service account \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the jobs service account with the
	// jobs RBAC role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		&rbac_v1.RoleBinding{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "jobs",
			},
			Subjects: []rbac_v1.Subject{
				rbac_v1.Subject{
					Kind:      "ServiceAccount",
					Name:      "jobs",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbac_v1.RoleRef{
				Kind: "Role",
				Name: "jobs",
			},
		},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create project secrets
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Create(
		&v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: project.ID,
				Labels: map[string]string{
					projectLabel: project.ID,
				},
			},
			StringData: project.Secrets,
		}); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project secret %q in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}

	return project, nil
}

func (s *scheduler) UpdateProject(project brignext.Project) error {
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Update(
		&v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: project.ID,
			},
			StringData: project.Secrets,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating project secret %q in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) DeleteProject(project brignext.Project) error {
	if err := s.kubeClient.CoreV1().Namespaces().Delete(
		project.Kubernetes.Namespace,
		&meta_v1.DeleteOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) CreateEvent(event brignext.Event) error {
	// Create a config map with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string                         `json:"id"`
			ProjectID  string                         `json:"projectID"`
			Source     string                         `json:"source"`
			Type       string                         `json:"type"`
			ShortTitle string                         `json:"shortTitle"`
			LongTitle  string                         `json:"longTitle"`
			Kubernetes brignext.EventKubernetesConfig `json:"kubernetes"`
		}{
			ID:         event.ID,
			ProjectID:  event.ProjectID,
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Kubernetes: *event.Kubernetes,
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(err, "error marshaling event %q", event.ID)
	}
	if _, err := s.kubeClient.CoreV1().ConfigMaps(
		event.Kubernetes.Namespace,
	).Create(
		&v1.ConfigMap{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: event.ID,
				Labels: map[string]string{
					projectLabel: event.ProjectID,
					eventLabel:   event.ID,
				},
			},
			Data: map[string]string{
				"event.json": string(eventJSON),
			},
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating event %q config map in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Create an event secret that is a point-in-time snapshot of the project's
	// secret
	projectSecret, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Get(event.ProjectID, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error finding existing project %q secret in namespace %q",
			event.ProjectID,
			event.Kubernetes.Namespace,
		)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Create(
		&v1.Secret{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: event.ID,
				Labels: map[string]string{
					projectLabel: event.ProjectID,
					eventLabel:   event.ID,
				},
			},
			Data: projectSecret.Data,
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating event %q secret in namespace %q",
			event.ProjectID,
			event.Kubernetes.Namespace,
		)
	}

	// For each of the event's workers, create a config map with worker details
	for workerName, worker := range event.Workers {
		workerJSON, err := json.MarshalIndent(
			struct {
				Name       string                   `json:"name"`
				Git        brignext.WorkerGitConfig `json:"git"`
				JobsConfig brignext.JobsConfig      `json:"jobsConfig"`
				LogLevel   brignext.LogLevel        `json:"logLevel"`
			}{
				Name:       workerName,
				Git:        worker.Git,
				JobsConfig: worker.JobsConfig,
				LogLevel:   worker.LogLevel,
			},
			"",
			"  ",
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"error marshaling worker %q of event %q to create a config map",
				workerName,
				event.ID,
			)
		}
		if _, err := s.kubeClient.CoreV1().ConfigMaps(
			event.Kubernetes.Namespace,
		).Create(
			&v1.ConfigMap{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", event.ID, strings.ToLower(workerName)),
					Labels: map[string]string{
						projectLabel: event.ProjectID,
						eventLabel:   event.ID,
						workerLabel:  workerName,
					},
				},
				Data: map[string]string{
					"worker.json": string(workerJSON),
				},
			},
		); err != nil {
			return errors.Wrapf(
				err,
				"error creating config map for worker %q of event %q",
				workerName,
				event.ID,
			)
		}
	}

	// Schedule all workers for asynchronous execution
	producer := redisMessaging.NewProducer(event.ProjectID, s.redisClient, nil)
	for workerName := range event.Workers {
		messageBody, err := json.Marshal(struct {
			Event  string `json:"event"`
			Worker string `json:"worker"`
		}{
			Event:  event.ID,
			Worker: workerName,
		})
		if err != nil {
			return errors.Wrapf(
				err,
				"error encoding execution task for event %q worker %q",
				event.ID,
				workerName,
			)
		}
		// TODO: Fix this
		// There's deliberately a short delay here to minimize the possibility of
		// the controller trying (and failing) to locate this new event before the
		// transaction on the store has become durable.
		if err := producer.Publish(
			messaging.NewDelayedMessage(messageBody, 5*time.Second),
		); err != nil {
			return errors.Wrapf(
				err,
				"error submitting execution task for event %q worker %q",
				event.ID,
				workerName,
			)
		}
	}

	return nil
}

func (s *scheduler) DeleteEvent(event brignext.Event) error {
	labels := map[string]string{
		eventLabel: event.ID,
	}
	if err := s.deletePodsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}
	if err := s.deleteConfigMapsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}
	if err := s.deleteSecretsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) DeleteWorker(
	event brignext.Event,
	workerName string,
) error {
	labels := map[string]string{
		eventLabel:  event.ID,
		workerLabel: workerName,
	}
	if err := s.deletePodsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q worker %q config maps in namespace %q",
			event.ID,
			workerName,
			event.Kubernetes.Namespace,
		)
	}
	if err := s.deleteConfigMapsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q worker %q config maps in namespace %q",
			event.ID,
			workerName,
			event.Kubernetes.Namespace,
		)
	}
	if err := s.deleteSecretsByLabels(
		event.Kubernetes.Namespace,
		labels,
	); err != nil {
		errors.Wrapf(
			err,
			"error deleting event %q worker %q secrets in namespace %q",
			event.ID,
			workerName,
			event.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) deletePodsByLabels(
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().Pods(namespace).DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}

func (s *scheduler) deleteConfigMapsByLabels(
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}

func (s *scheduler) deleteSecretsByLabels(
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().Secrets(namespace).DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}
