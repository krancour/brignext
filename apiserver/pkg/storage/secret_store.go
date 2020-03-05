package storage

import "github.com/krancour/brignext"

type SecretStore interface {
	CreateProjectSecrets(namespace, id string, secrets map[string]string) error
	UpdateProjectSecrets(namespace, id string, secrets map[string]string) error
	DeleteProjectSecrets(namespace, id string) error

	// TODO: Move these. They're not secrets! Or should they be???
	CreateEventConfigMap(brignext.Event) error
	// TODO: This isn't called anywhere yet. This will be critical cleannup.
	DeleteEventConfigMap(namespace, id string) error

	CreateEventSecrets(namespace, projectID, eventID string) error
	// TODO: This isn't called anywhere yet. This will be critical cleannup.
	DeleteEventSecrets(namespace, id string) error

	// TODO: Move these. They're not secrets! Or should they be???
	CreateWorkerConfigMap(
		namespace string,
		projectID string,
		eventID string,
		workerName string,
		worker brignext.Worker,
	) error
	// TODO: This isn't called anywhere yet. This will be critical cleannup.
	DeleteWorkerConfigMap(namespace, eventID, workerName string) error
}
