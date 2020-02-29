package storage

import "github.com/krancour/brignext"

type SecretStore interface {
	CreateProjectSecrets(id, namespace string, secrets map[string]string) error
	UpdateProjectSecrets(id, namespace string, secrets map[string]string) error
	DeleteProjectSecrets(id, namespace string) error

	// TODO: Move these. They're not secrets!
	CreateEventConfigMap(brignext.Event) error
	DeleteEventConfigMap(id, namespace string) error

	CreateEventSecrets(projectID, namespace, eventID string) error
	DeleteEventSecrets(id, namespace string) error

	// TODO: Move these. They're not secrets!
	CreateWorkerConfigMap(
		event brignext.Event,
		workerName string,
		worker brignext.Worker,
	) error
	DeleteWorkerConfigMap(eventID, workerName string) error
}
