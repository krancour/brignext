package events

import (
	"context"
	"encoding/json"
	"fmt"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	myk8s "github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

type Scheduler interface {
	PreCreate(
		ctx context.Context,
		project brignext.Project,
		event brignext.Event,
	) (brignext.Event, error)
	Create(
		ctx context.Context,
		project brignext.Project,
		event brignext.Event,
	) error
	Delete(context.Context, brignext.Event) error

	StartWorker(ctx context.Context, event brignext.Event) error

	StartJob(
		ctx context.Context,
		event brignext.Event,
		jobName string,
	) error
}

type scheduler struct {
	config              Config
	eventsSenderFactory SenderFactory
	kubeClient          *kubernetes.Clientset
}

func NewScheduler(
	config Config,
	eventsSenderFactory SenderFactory,
	kubeClient *kubernetes.Clientset,
) Scheduler {
	return &scheduler{
		config:              config,
		eventsSenderFactory: eventsSenderFactory,
		kubeClient:          kubeClient,
	}
}

func (s *scheduler) PreCreate(
	ctx context.Context,
	proj brignext.Project,
	event brignext.Event,
) (brignext.Event, error) {
	// Fill in scheduler-specific details
	event.Kubernetes = proj.Kubernetes
	event.Worker.Spec.Kubernetes = proj.Spec.WorkerTemplate.Kubernetes
	return event, nil
}

func (s *scheduler) Create(
	ctx context.Context,
	proj brignext.Project,
	event brignext.Event,
) error {
	// Get the project's secrets
	projectSecretsSecret, err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
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
		ID         string                     `json:"id"`
		Kubernetes *brignext.KubernetesConfig `json:"kubernetes"`
		Secrets    map[string]string          `json:"secrets"`
	}

	type worker struct {
		APIToken             string            `json:"apiToken"`
		LogLevel             brignext.LogLevel `json:"logLevel"`
		ConfigFilesDirectory string            `json:"configFilesDirectory"`
		DefaultConfigFiles   map[string]string `json:"defaultConfigFiles" bson:"defaultConfigFiles"` // nolint: lll
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
				Kubernetes: event.Kubernetes,
				Secrets:    secrets,
			},
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Payload:    event.Payload,
			Worker: worker{
				APIToken:             event.Worker.Token,
				LogLevel:             event.Worker.Spec.LogLevel,
				ConfigFilesDirectory: event.Worker.Spec.ConfigFilesDirectory,
				DefaultConfigFiles:   event.Worker.Spec.DefaultConfigFiles,
			},
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(err, "error marshaling event %q", event.ID)
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
					myk8s.ComponentLabel: "event",
					myk8s.ProjectLabel:   event.ProjectID,
					myk8s.EventLabel:     event.ID,
				},
			},
			Type: corev1.SecretType("brignext.io/event"),
			Data: data,
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Schedule event's worker for asynchronous execution
	eventSender, err := s.eventsSenderFactory.NewSender(event.ProjectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating event sender for event %q",
			event.ID,
		)
	}
	if err := eventSender.Send(ctx, event.ID); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for event %q worker",
			event.ID,
		)
	}

	return nil
}

func (s *scheduler) Delete(
	ctx context.Context,
	event brignext.Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		myk8s.EventLabel,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this event
	if err := s.deletePodsByLabelSelector(
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
	if err := s.deletePersistentVolumeClaimsByLabelSelector(
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
	if err := s.deleteConfigMapsByLabelSelector(
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
	if err := s.deleteSecretsByLabelSelector(
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

func (s *scheduler) StartWorker(
	ctx context.Context,
	event brignext.Event,
) error {
	if err := s.createWorkspacePVC(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error creating workspace for event %q worker",
			event.ID,
		)
	}
	if err := s.createWorkerPod(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q worker",
			event.ID,
		)
	}
	return nil
}

// TODO: Implement this
func (s *scheduler) StartJob(
	ctx context.Context,
	event brignext.Event,
	jobName string,
) error {
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
