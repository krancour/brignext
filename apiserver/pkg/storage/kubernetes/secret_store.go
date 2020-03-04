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
	namespace string,
	id string,
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
	namespace string,
	id string,
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
	namespace string,
	id string,
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
			ID         string             `json:"id"`
			ProjectID  string             `json:"projectID"`
			Provider   string             `json:"provider"`
			Type       string             `json:"type"`
			ShortTitle string             `json:"shortTitle"`
			LongTitle  string             `json:"longTitle"`
			Git        brignext.GitConfig `json:"git"`
		}{
			ID:         event.ID,
			ProjectID:  event.ProjectID,
			Provider:   event.Provider,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Git:        event.Git,
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
func (s *secretStore) DeleteEventConfigMap(namespace, id string) error {
	return nil
}

func (s *secretStore) CreateEventSecrets(
	namespace string,
	projectID string,
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

func (s *secretStore) DeleteEventSecrets(namespace, id string) error {
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
	namespace string,
	projectID string,
	eventID string,
	workerName string,
	worker brignext.Worker,
) error {
	configMapsClient :=
		s.kubeClient.CoreV1().ConfigMaps(namespace)

	workerJSON, err := json.MarshalIndent(
		struct {
			Name     string              `json:"name"`
			Git      brignext.GitConfig  `json:"git"`
			Jobs     brignext.JobsConfig `json:"jobs"`
			LogLevel brignext.LogLevel   `json:"logLevel"`
		}{
			Name:     workerName,
			Git:      worker.Git,
			Jobs:     worker.Jobs,
			LogLevel: worker.LogLevel,
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling worker %q of event %q to create a config map",
			workerName,
			eventID,
		)
	}

	if _, err := configMapsClient.Create(&v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", eventID, strings.ToLower(workerName)),
			Namespace: namespace,
		},
		Data: map[string]string{
			"worker.json": string(workerJSON),
		},
	}); err != nil {
		return errors.Wrapf(
			err,
			"error creating config map for worker %q of event %q",
			workerName,
			eventID,
		)
	}

	return nil
}

// TODO: Implement this
func (s *secretStore) DeleteWorkerConfigMap(
	namespace string,
	eventID string,
	workerName string,
) error {
	return nil
}
