package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/internal/messaging"
	redisMessaging "github.com/krancour/brignext/v2/internal/messaging/redis"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

// TODO: These might have been duplicated in a few places
const (
	componentLabel = "brignext.io/component"
	projectLabel   = "brignext.io/project"
	eventLabel     = "brignext.io/event"
)

type Scheduler interface {
	// TODO: Add a PreCreate func!
	Create(
		ctx context.Context,
		project brignext.Project,
		event brignext.Event,
	) (brignext.Event, error)
	Delete(context.Context, brignext.EventReference) error

	CheckHealth(context.Context) error
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

func (s *scheduler) Create(
	ctx context.Context,
	proj brignext.Project,
	event brignext.Event,
) (brignext.Event, error) {
	// Fill in scheduler-specific details
	event.Kubernetes = proj.Kubernetes
	event.Worker.Kubernetes = proj.Spec.Worker.Kubernetes

	// Get the project's secrets
	projectSecretsSecret, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error finding secret \"project-secrets\" in namespace %q",
			event.Kubernetes.Namespace,
		)
	}
	secrets := map[string]string{}
	for key, value := range projectSecretsSecret.Data {
		secrets[key] = string(value)
	}

	type project struct {
		ID         string                    `json:"id"`
		Kubernetes brignext.KubernetesConfig `json:"kubernetes"`
		Secrets    map[string]string         `json:"secrets"`
	}

	type worker struct {
		Git                  brignext.WorkerGitConfig `json:"git"`
		Jobs                 brignext.JobsSpec        `json:"jobs"`
		LogLevel             brignext.LogLevel        `json:"logLevel"`
		ConfigFilesDirectory string                   `json:"configFilesDirectory"`
		DefaultConfigFiles   map[string]string        `json:"defaultConfigFiles" bson:"defaultConfigFiles"` // nolint: lll
	}

	// Create a secret with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string  `json:"id"`
			Project    project `json:"project"`
			Source     string  `json:"source"`
			Type       string  `json:"type"`
			ShortTitle string  `json:"shortTitle"`
			LongTitle  string  `json:"longTitle"`
			Payload    string  `json:"payload"`
			Worker     worker  `json:"worker"`
		}{
			ID: event.ID,
			Project: project{
				ID:         event.ProjectID,
				Kubernetes: *event.Kubernetes,
				Secrets:    secrets,
			},
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Payload:    event.Payload,
			Worker: worker{
				Git:                  event.Worker.Git,
				Jobs:                 event.Worker.Jobs,
				LogLevel:             event.Worker.LogLevel,
				ConfigFilesDirectory: event.Worker.ConfigFilesDirectory,
				DefaultConfigFiles:   event.Worker.DefaultConfigFiles,
			},
		},
		"",
		"  ",
	)
	if err != nil {
		return event, errors.Wrapf(err, "error marshaling event %q", event.ID)
	}

	data := map[string][]byte{}
	data["event.json"] = eventJSON
	data["gitSSHKey"] = projectSecretsSecret.Data["gitSSHKey"]
	data["gitSSHCert"] = projectSecretsSecret.Data["gitSSHCert"]

	if _, err = s.kubeClient.CoreV1().Secrets(
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
			Data: data,
		},
		metav1.CreateOptions{},
	); err != nil {
		return event, errors.Wrapf(
			err,
			"error creating secret %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Schedule event's worker for asynchronous execution
	producer := redisMessaging.NewProducer(event.ProjectID, s.redisClient, nil)
	messageBody, err := json.Marshal(struct {
		Event string `json:"event"`
	}{
		Event: event.ID,
	})
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error encoding execution task for event %q worker",
			event.ID,
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
			"error submitting execution task for event %q worker",
			event.ID,
		)
	}

	return event, nil
}

func (s *scheduler) Delete(
	ctx context.Context,
	eventRef brignext.EventReference,
) error {
	matchesEvent, _ := labels.NewRequirement(
		eventLabel,
		selection.Equals,
		[]string{eventRef.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this event
	if err := s.deletePodsByLabelSelector(
		ctx,
		eventRef.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			eventRef.ID,
			eventRef.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related to this event
	if err := s.deletePersistentVolumeClaimsByLabelSelector(
		ctx,
		eventRef.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			eventRef.ID,
			eventRef.Kubernetes.Namespace,
		)
	}

	// Delete all config maps related to this event. BrigNext itself doesn't
	// create any, but we're not discounting the possibility that a worker or job
	// might create some. We are, of course, assuming that anything created by a
	// worker or job is labeled appropriately.
	if err := s.deleteConfigMapsByLabelSelector(
		ctx,
		eventRef.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			eventRef.ID,
			eventRef.Kubernetes.Namespace,
		)
	}

	// Delete all secrets related to this event
	if err := s.deleteSecretsByLabelSelector(
		ctx,
		eventRef.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			eventRef.ID,
			eventRef.Kubernetes.Namespace,
		)
	}

	return nil
}

func (s *scheduler) CheckHealth(context.Context) error {
	// TODO: Ping the Kubernetes apiserver
	if err := s.redisClient.Ping().Err(); err != nil {
		return errors.Wrap(err, "error pinging redis")
	}
	return nil
}

func (s *scheduler) deletePodsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return s.kubeClient.CoreV1().Pods(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (s *scheduler) deletePersistentVolumeClaimsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return s.kubeClient.CoreV1().PersistentVolumeClaims(
		namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (s *scheduler) deleteConfigMapsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return s.kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (s *scheduler) deleteSecretsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return s.kubeClient.CoreV1().Secrets(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}
