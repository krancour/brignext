package storage

type SecretStore interface {
	CreateProjectSecrets(id, namespace string, secrets map[string]string) error
	UpdateProjectSecrets(id, namespace string, secrets map[string]string) error
	DeleteProjectSecrets(id, namespace string) error

	CreateEventSecrets(projectID, namespace, eventID string) error
	DeleteEventSecrets(id, namespace string) error
}
