package kubernetes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/storage"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type secretStore struct {
	kubeClient *kubernetes.Clientset
}

func NewSecretStore(kubeClient *kubernetes.Clientset) storage.SecretStore {
	return &secretStore{
		kubeClient: kubeClient,
	}
}

func (s *secretStore) CreateProjectSecrets(
	id string,
	namespace string,
	secrets map[string]string,
) error {
	secretsClient := s.kubeClient.CoreV1().Secrets(namespace)

	if _, err := secretsClient.Create(&v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
		StringData: secrets,
	}); err != nil {
		return errors.Wrapf(
			err,
			"error creating project secret %q in namespace %q",
			id,
			namespace,
		)
	}

	return nil
}

func (s *secretStore) UpdateProjectSecrets(
	id string,
	namespace string,
	secrets map[string]string,
) error {
	secretsClient := s.kubeClient.CoreV1().Secrets(namespace)

	if _, err := secretsClient.Update(&v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
		StringData: secrets,
	}); err != nil {
		return errors.Wrapf(
			err,
			"error updating project secret %q in namespace %q",
			id,
			namespace,
		)
	}

	return nil
}

func (s *secretStore) DeleteProjectSecrets(
	id string,
	namespace string,
) error {
	secretsClient := s.kubeClient.CoreV1().Secrets(namespace)

	if err := secretsClient.Delete(id, &meta_v1.DeleteOptions{}); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project secret %q in namespace %q",
			id,
			namespace,
		)
	}

	return nil
}

func (s *secretStore) CreateEventConfigMap(event brignext.Event) error {
	configMapsClient :=
		s.kubeClient.CoreV1().ConfigMaps(event.Kubernetes.Namespace)

	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string              `json:"id,omitempty"`
			ProjectID  string              `json:"projectID,omitempty"`
			Provider   string              `json:"provider,omitempty"`
			Type       string              `json:"type,omitempty"`
			ShortTitle string              `json:"shortTitle,omitempty"`
			LongTitle  string              `json:"longTitle,omitempty"`
			Git        *brignext.GitConfig `json:"git,omitempty"`
			Namespace  string              `json:"namespace,omitempty"`
		}{
			ID:         event.ID,
			ProjectID:  event.ProjectID,
			Provider:   event.Provider,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Git:        event.Git,
			Namespace:  event.Kubernetes.Namespace,
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling event %q to create a config map",
			event.ID,
		)
	}

	if _, err := configMapsClient.Create(&v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      event.ID,
			Namespace: event.Kubernetes.Namespace,
		},
		Data: map[string]string{
			"event.json": string(eventJSON),
		},
	}); err != nil {
		return errors.Wrapf(err, "error creating event %q config map", event.ID)
	}

	return nil
}

// TODO: Implement this
func (s *secretStore) DeleteEventConfigMap(id, namespace string) error {
	return nil
}

func (s *secretStore) CreateEventSecrets(
	projectID string,
	namespace string,
	eventID string,
) error {
	secretsClient := s.kubeClient.CoreV1().Secrets(namespace)

	projectSecret, err := secretsClient.Get(projectID, meta_v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error checking for existing project secret %q in namespace %q",
			projectID,
			namespace,
		)
	}

	if _, err := secretsClient.Create(&v1.Secret{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      eventID,
			Namespace: namespace,
		},
		Data: projectSecret.Data,
	}); err != nil {
		return errors.Wrapf(
			err,
			"error creating event secret %q in namespace %q",
			eventID,
			namespace,
		)
	}

	return nil
}

func (s *secretStore) DeleteEventSecrets(id, namespace string) error {
	secretsClient := s.kubeClient.CoreV1().Secrets(namespace)

	if err := secretsClient.Delete(id, &meta_v1.DeleteOptions{}); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event secret %q in namespace %q",
			id,
			namespace,
		)
	}

	return nil
}

func (s *secretStore) CreateWorkerConfigMap(
	event brignext.Event,
	workerName string,
	worker brignext.Worker,
) error {
	configMapsClient :=
		s.kubeClient.CoreV1().ConfigMaps(event.Kubernetes.Namespace)

	workerJSON, err := json.MarshalIndent(
		struct {
			Name       string                           `json:"name,omitempty"`
			Git        *brignext.GitConfig              `json:"git,omitempty"`
			Kubernetes *brignext.WorkerKubernetesConfig `json:"kubernetes,omitempty"`
			Jobs       *brignext.JobsConfig             `json:"jobs,omitempty"`
			LogLevel   brignext.LogLevel                `json:"logLevel,omitempty"`
		}{
			Name:       workerName,
			Git:        worker.Git,
			Kubernetes: worker.Kubernetes,
			Jobs:       worker.Jobs,
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

	if _, err := configMapsClient.Create(&v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", event.ID, strings.ToLower(workerName)),
			Namespace: event.Kubernetes.Namespace,
		},
		Data: map[string]string{
			"worker.json": string(workerJSON),
		},
	}); err != nil {
		return errors.Wrapf(
			err,
			"error creating config map for worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	return nil
}

// TODO: Implement this
func (s *secretStore) DeleteWorkerConfigMap(eventID, workerName string) error {
	return nil
}
