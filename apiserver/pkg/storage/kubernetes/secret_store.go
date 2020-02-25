package kubernetes

import (
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
