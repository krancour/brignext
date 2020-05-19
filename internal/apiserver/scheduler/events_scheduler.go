package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/common/messaging"
	redisMessaging "github.com/krancour/brignext/v2/internal/common/messaging/redis" // nolint: lll
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

type EventsScheduler interface {
	Create(
		ctx context.Context,
		project brignext.Project,
		event brignext.Event,
	) (brignext.Event, error)
	Get(ctx context.Context, event brignext.Event) (brignext.Event, error)
	Cancel(ctx context.Context, event brignext.Event) error
	Delete(ctx context.Context, event brignext.Event) error
}

type eventsScheduler struct {
	redisClient *redis.Client
	kubeClient  *kubernetes.Clientset
}

func NewEventsScheduler(
	redisClient *redis.Client,
	kubeClient *kubernetes.Clientset,
) EventsScheduler {
	return &eventsScheduler{
		redisClient: redisClient,
		kubeClient:  kubeClient,
	}
}

func (e *eventsScheduler) Create(
	ctx context.Context,
	project brignext.Project,
	event brignext.Event,
) (brignext.Event, error) {
	// Fill in scheduler-specific details
	event.Kubernetes = project.Kubernetes
	event.Worker.Kubernetes = project.Spec.Worker.Kubernetes

	// Create a secret with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string                    `json:"id"`
			ProjectID  string                    `json:"projectID"`
			Source     string                    `json:"source"`
			Type       string                    `json:"type"`
			ShortTitle string                    `json:"shortTitle"`
			LongTitle  string                    `json:"longTitle"`
			Kubernetes brignext.KubernetesConfig `json:"kubernetes"`
			Payload    string                    `json:"payload"`
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
	if _, err = e.kubeClient.CoreV1().Secrets(
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

	// Get the project's secrets
	projectSecretsSecret, err := e.kubeClient.CoreV1().Secrets(
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

	workerJSON, err := json.MarshalIndent(
		struct {
			Git                  brignext.WorkerGitConfig `json:"git"`
			Jobs                 brignext.JobsSpec        `json:"jobs"`
			LogLevel             brignext.LogLevel        `json:"logLevel"`
			Secrets              map[string]string        `json:"secrets"`
			ConfigFilesDirectory string                   `json:"configFilesDirectory"` // nolint: lll
		}{
			Git:                  event.Worker.Git,
			Jobs:                 event.Worker.Jobs,
			LogLevel:             event.Worker.LogLevel,
			Secrets:              secrets,
			ConfigFilesDirectory: event.Worker.ConfigFilesDirectory,
		},
		"",
		"  ",
	)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error marshaling worker for event %q to create a worker secret",
			event.ID,
		)
	}
	data := map[string][]byte{}
	for filename, contents := range event.Worker.DefaultConfigFiles {
		data[filename] = []byte(contents)
	}
	data["worker.json"] = workerJSON
	data["gitSSHKey"] = projectSecretsSecret.Data["gitSSHKey"]
	data["gitSSHCert"] = projectSecretsSecret.Data["gitSSHCert"]
	if _, err = e.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("worker-%s", event.ID),
				Labels: map[string]string{
					componentLabel: "worker",
					projectLabel:   event.ProjectID,
					eventLabel:     event.ID,
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
			fmt.Sprintf("worker-%s", event.ID),
			event.Kubernetes.Namespace,
		)
	}

	// Schedule event's worker for asynchronous execution
	producer := redisMessaging.NewProducer(event.ProjectID, e.redisClient, nil)
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

func (e *eventsScheduler) Get(
	ctx context.Context,
	event brignext.Event,
) (brignext.Event, error) {
	eventSecret, err := e.kubeClient.CoreV1().Secrets(
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

func (e *eventsScheduler) Cancel(
	ctx context.Context,
	event brignext.Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		eventLabel,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this event
	if err := e.deletePodsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related this this event
	if err := e.deletePersistentVolumeClaimsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all config maps related to this event. BrigNext itself doesn't
	// create any, but we're not discounting the possibility that a worker or job
	// might create some. We are, of course, assuming that anything created by a
	// worker or job is labeled appropriately.
	if err := e.deleteConfigMapsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete most secrets related to the event, being careful not to delete the
	// secret that contains the event's payload. As long as the event exists,
	// we'll want this to stick around.
	isntEventSecret, _ := labels.NewRequirement(
		componentLabel,
		selection.NotEquals,
		[]string{"event"},
	)
	labelSelector = labelSelector.Add(*isntEventSecret)
	if err := e.deleteSecretsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	return nil
}

func (e *eventsScheduler) Delete(
	ctx context.Context,
	event brignext.Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		eventLabel,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this event
	if err := e.deletePodsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related to this event
	if err := e.deletePersistentVolumeClaimsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all config maps related to this event. BrigNext itself doesn't
	// create any, but we're not discounting the possibility that a worker or job
	// might create some. We are, of course, assuming that anything created by a
	// worker or job is labeled appropriately.
	if err := e.deleteConfigMapsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all secrets related to this event
	if err := e.deleteSecretsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	return nil
}

func (e *eventsScheduler) deletePodsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().Pods(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deletePersistentVolumeClaimsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().PersistentVolumeClaims(
		namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deleteConfigMapsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deleteSecretsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().Secrets(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}
