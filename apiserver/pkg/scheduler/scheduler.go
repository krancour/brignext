package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/apiserver/pkg/crypto"
	"github.com/krancour/brignext/v2/pkg/messaging"
	redisMessaging "github.com/krancour/brignext/v2/pkg/messaging/redis"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	CreateProject(
		ctx context.Context,
		project brignext.Project,
	) (brignext.Project, error)
	UpdateProject(
		ctx context.Context,
		project brignext.Project,
	) (brignext.Project, error)
	DeleteProject(
		ctx context.Context,
		project brignext.Project,
	) error

	GetSecrets(
		ctx context.Context,
		project brignext.Project,
		workerName string,
	) (map[string]string, error)
	SetSecrets(
		ctx context.Context,
		project brignext.Project,
		workerName string,
		secrets map[string]string,
	) error
	UnsetSecrets(
		ctx context.Context,
		project brignext.Project,
		workerName string,
		keys []string,
	) error

	CreateEvent(
		ctx context.Context,
		project brignext.Project,
		event brignext.Event,
	) (brignext.Event, error)
	GetEvent(ctx context.Context, event brignext.Event) (brignext.Event, error)
	DeleteEvent(ctx context.Context, event brignext.Event) error

	DeleteWorker(
		ctx context.Context,
		event brignext.Event,
		workerName string,
	) error
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
	ctx context.Context,
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
		ctx,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Kubernetes.Namespace,
			},
		},
		metav1.CreateOptions{},
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
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Rules: []rbacv1.PolicyRule{
				rbacv1.PolicyRule{
					APIGroups: []string{""},
					Resources: []string{"configmaps", "secrets", "pods", "pods/log"},
					Verbs:     []string{"create", "get", "list", "watch"},
				},
			},
		},
		metav1.CreateOptions{},
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
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
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
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "workers",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role for use by all of the project's jobs
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			// TODO: Add the correct rules here
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
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
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
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
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "jobs",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a secret for each worker config
	for workerName := range project.WorkerConfigs {
		if _, err := s.kubeClient.CoreV1().Secrets(
			project.Kubernetes.Namespace,
		).Create(
			ctx,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("worker-config-%s", workerName),
					Labels: map[string]string{
						componentLabel: "worker-config",
						projectLabel:   project.ID,
						workerLabel:    workerName,
					},
				},
				Type: corev1.SecretType("brignext.io/worker-config"),
			},
			metav1.CreateOptions{},
		); err != nil {
			return project, errors.Wrapf(
				err,
				"error creating secret %q in namespace %q",
				fmt.Sprintf("worker-config-%s", workerName),
				project.Kubernetes.Namespace,
			)
		}
	}

	return project, nil
}

func (s *scheduler) UpdateProject(
	ctx context.Context,
	project brignext.Project,
) (brignext.Project, error) {
	// This is a no-op
	return project, nil
}

func (s *scheduler) DeleteProject(
	ctx context.Context,
	project brignext.Project,
) error {
	if err := s.kubeClient.CoreV1().Namespaces().Delete(
		ctx,
		project.Kubernetes.Namespace,
		metav1.DeleteOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) GetSecrets(
	ctx context.Context,
	project brignext.Project,
	workerName string,
) (map[string]string, error) {
	secret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(
		ctx,
		fmt.Sprintf("worker-config-%s", workerName),
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving secret %q in namespace %q",
			fmt.Sprintf("worker-config-%s", workerName),
			project.Kubernetes.Namespace,
		)
	}
	secrets := map[string]string{}
	for key := range secret.Data {
		secrets[key] = "*** REDACTED ***"
	}
	return secrets, nil
}

func (s *scheduler) SetSecrets(
	ctx context.Context,
	project brignext.Project,
	workerName string,
	secrets map[string]string,
) error {
	secret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(
		ctx,
		fmt.Sprintf("worker-config-%s", workerName),
		metav1.GetOptions{},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving secret %q in namespace %q",
			fmt.Sprintf("worker-config-%s", workerName),
			project.Kubernetes.Namespace,
		)
	}
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	for key, value := range secrets {
		secret.Data[key] = []byte(value)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(
			err,
			"error updating project secret %q in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) UnsetSecrets(
	ctx context.Context,
	project brignext.Project,
	workerName string,
	keys []string,
) error {
	secret, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(
		ctx,
		fmt.Sprintf("worker-config-%s", workerName),
		metav1.GetOptions{},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving secret %q in namespace %q",
			fmt.Sprintf("worker-config-%s", workerName),
			project.Kubernetes.Namespace,
		)
	}
	for _, key := range keys {
		delete(secret.Data, key)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return errors.Wrapf(
			err,
			"error updating project secret %q in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *scheduler) CreateEvent(
	ctx context.Context,
	project brignext.Project,
	event brignext.Event,
) (brignext.Event, error) {
	// Fill in scheduler-specific details
	event.Kubernetes = &brignext.EventKubernetesConfig{
		Namespace: project.Kubernetes.Namespace,
	}
	for workerName, worker := range event.Workers {
		worker.Kubernetes = project.WorkerConfigs[workerName].Kubernetes
		event.Workers[workerName] = worker
	}

	// Create a secret with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string                         `json:"id"`
			ProjectID  string                         `json:"projectID"`
			Source     string                         `json:"source"`
			Type       string                         `json:"type"`
			ShortTitle string                         `json:"shortTitle"`
			LongTitle  string                         `json:"longTitle"`
			Kubernetes brignext.EventKubernetesConfig `json:"kubernetes"`
			Payload    string                         `json:"payload"`
		}{
			ID:         event.ID,
			ProjectID:  event.ProjectID,
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Kubernetes: *event.Kubernetes,
			Payload:    event.Payload,
		},
		"",
		"  ",
	)
	if err != nil {
		return event, errors.Wrapf(err, "error marshaling event %q", event.ID)
	}
	if _, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("event-%s", event.ID),
				Labels: map[string]string{
					componentLabel: "event",
					projectLabel:   event.ProjectID,
					eventLabel:     event.ID,
				},
			},
			Type: corev1.SecretType("brignext.io/event"),
			StringData: map[string]string{
				"event.json": string(eventJSON),
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return event, errors.Wrapf(
			err,
			"error creating config map %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// For each of the event's workers, create a secret with worker details.
	for workerName, worker := range event.Workers {
		// Get the worker config's secrets
		workerConfigSecret, err := s.kubeClient.CoreV1().Secrets(
			event.Kubernetes.Namespace,
		).Get(
			ctx,
			fmt.Sprintf("worker-config-%s", workerName),
			metav1.GetOptions{},
		)
		if err != nil {
			return event, errors.Wrapf(
				err,
				"error finding secret %q in namespace %q",
				fmt.Sprintf("worker-config-%s", workerName),
				event.Kubernetes.Namespace,
			)
		}
		secrets := map[string]string{}
		for key, value := range workerConfigSecret.Data {
			secrets[key] = string(value)
		}

		workerJSON, err := json.MarshalIndent(
			struct {
				Name                 string                   `json:"name"`
				Git                  brignext.WorkerGitConfig `json:"git"`
				JobsConfig           brignext.JobsConfig      `json:"jobsConfig"`
				LogLevel             brignext.LogLevel        `json:"logLevel"`
				Secrets              map[string]string        `json:"secrets"`
				ConfigFilesDirectory string                   `json:"configFilesDirectory"` // nolint: lll
			}{
				Name:                 workerName,
				Git:                  worker.Git,
				JobsConfig:           worker.JobsConfig,
				LogLevel:             worker.LogLevel,
				Secrets:              secrets,
				ConfigFilesDirectory: worker.ConfigFilesDirectory,
			},
			"",
			"  ",
		)
		if err != nil {
			return event, errors.Wrapf(
				err,
				"error marshaling worker %q of event %q to create a config map",
				workerName,
				event.ID,
			)
		}
		data := map[string][]byte{}
		for filename, contents := range worker.DefaultConfigFiles {
			data[filename] = []byte(contents)
		}
		data["worker.json"] = workerJSON
		data["gitSSHKey"] = workerConfigSecret.Data["gitSSHKey"]
		data["gitSSHCert"] = workerConfigSecret.Data["gitSSHCert"]
		if _, err := s.kubeClient.CoreV1().Secrets(
			event.Kubernetes.Namespace,
		).Create(
			ctx,
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("worker-%s-%s", event.ID, workerName),
					Labels: map[string]string{
						componentLabel: "worker",
						projectLabel:   event.ProjectID,
						eventLabel:     event.ID,
						workerLabel:    workerName,
					},
				},
				Type: corev1.SecretType("brignext.io/worker"),
				Data: data,
			},
			metav1.CreateOptions{},
		); err != nil {
			return event, errors.Wrapf(
				err,
				"error creating secret %q in namespace %q",
				fmt.Sprintf("%s-%s", event.ID, workerName),
				event.Kubernetes.Namespace,
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
			return event, errors.Wrapf(
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
			return event, errors.Wrapf(
				err,
				"error submitting execution task for event %q worker %q",
				event.ID,
				workerName,
			)
		}
	}

	return event, nil
}

func (s *scheduler) GetEvent(
	ctx context.Context,
	event brignext.Event,
) (brignext.Event, error) {
	eventSecret, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Get(
		ctx,
		fmt.Sprintf("event-%s", event.ID),
		metav1.GetOptions{},
	)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error finding secret %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}
	eventStruct := struct {
		Payload string `json:"payload"`
	}{}
	if err := json.Unmarshal(
		eventSecret.Data["event.json"],
		&eventStruct,
	); err != nil {
		return event, errors.Wrapf(
			err,
			"error unmarshaling event from secret %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}
	event.Payload = eventStruct.Payload
	return event, nil
}

func (s *scheduler) DeleteEvent(
	ctx context.Context,
	event brignext.Event,
) error {
	labels := map[string]string{
		eventLabel: event.ID,
	}
	if err := s.deletePodsByLabels(
		ctx,
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
		ctx,
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
		ctx,
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
	ctx context.Context,
	event brignext.Event,
	workerName string,
) error {
	labels := map[string]string{
		eventLabel:  event.ID,
		workerLabel: workerName,
	}
	if err := s.deletePodsByLabels(
		ctx,
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
		ctx,
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
		ctx,
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
	ctx context.Context,
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().Pods(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}

func (s *scheduler) deleteConfigMapsByLabels(
	ctx context.Context,
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}

func (s *scheduler) deleteSecretsByLabels(
	ctx context.Context,
	namespace string,
	labelsMap map[string]string,
) error {
	return s.kubeClient.CoreV1().Secrets(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(labelsMap).String(),
		},
	)
}
