package service

import (
	"context"
	"time"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/crypto"
	"github.com/krancour/brignext/apiserver/pkg/scheduler"
	"github.com/krancour/brignext/apiserver/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Service interface {
	CreateUser(context.Context, brignext.User) error
	GetUsers(context.Context) ([]brignext.User, error)
	GetUser(context.Context, string) (brignext.User, error)
	LockUser(context.Context, string) error
	UnlockUser(context.Context, string) error

	CreateRootSession(context.Context) (string, error)
	CreateUserSession(context.Context) (string, string, error)
	GetSessionByOAuth2State(context.Context, string) (brignext.Session, error)
	GetSessionByToken(context.Context, string) (brignext.Session, error)
	AuthenticateSession(
		ctx context.Context,
		sessionID string,
		userID string,
	) error
	DeleteSession(context.Context, string) error
	DeleteSessionsByUser(context.Context, string) (int64, error)

	CreateServiceAccount(context.Context, brignext.ServiceAccount) (string, error)
	GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error)
	GetServiceAccount(context.Context, string) (brignext.ServiceAccount, error)
	GetServiceAccountByToken(
		context.Context,
		string,
	) (brignext.ServiceAccount, error)
	LockServiceAccount(context.Context, string) error
	UnlockServiceAccount(context.Context, string) (string, error)

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) ([]brignext.Project, error)
	GetProject(context.Context, string) (brignext.Project, error)
	UpdateProject(context.Context, brignext.Project) error
	DeleteProject(context.Context, string) error

	CreateEvent(context.Context, brignext.Event) (string, error)
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, error)
	UpdateEventStatus(
		ctx context.Context,
		id string,
		status brignext.EventStatus,
	) error
	UpdateEventWorkerStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		status brignext.WorkerStatus,
	) error
	DeleteEvent(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteProcessing bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deletePending bool,
		deleteProcessing bool,
	) (int64, error)
}

type service struct {
	store       storage.Store
	secretStore storage.SecretStore
	scheduler   scheduler.Scheduler
	logStore    storage.LogStore
}

func NewService(
	store storage.Store,
	secretStore storage.SecretStore,
	scheduler scheduler.Scheduler,
	logStore storage.LogStore,
) Service {
	return &service{
		store:       store,
		secretStore: secretStore,
		scheduler:   scheduler,
		logStore:    logStore,
	}
}

func (s *service) CreateUser(ctx context.Context, user brignext.User) error {
	if _, err := s.store.GetUser(ctx, user.ID); err != nil {
		if _, ok := err.(*brignext.ErrUserNotFound); !ok {
			return errors.Wrapf(err, "error searching for existing user %q", user.ID)
		}
	} else {
		return &brignext.ErrUserIDConflict{user.ID}
	}

	user.FirstSeen = time.Now()
	if err := s.store.CreateUser(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}

	return nil
}

func (s *service) GetUsers(ctx context.Context) ([]brignext.User, error) {
	users, err := s.store.GetUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users from store")
	}
	return users, nil
}

func (s *service) GetUser(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		return user, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, nil
}

func (s *service) LockUser(ctx context.Context, id string) error {
	return s.store.DoTx(ctx, func(ctx context.Context) error {

		var err error
		if err = s.store.LockUser(ctx, id); err != nil {
			return errors.Wrapf(err, "error locking user %q in store", id)
		}

		if _, err := s.store.DeleteSessionsByUser(ctx, id); err != nil {
			return errors.Wrapf(
				err,
				"error removing sessions for user %q from store",
				id,
			)
		}

		return nil
	})
}

func (s *service) UnlockUser(ctx context.Context, id string) error {
	if err := s.store.UnlockUser(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

func (s *service) CreateRootSession(ctx context.Context) (string, error) {
	token := crypto.NewToken(256)
	now := time.Now()
	session := brignext.Session{
		ID:            uuid.NewV4().String(),
		Root:          true,
		HashedToken:   crypto.ShortSHA("", token),
		Authenticated: true,
		Expires:       now.Add(10 * time.Minute),
		Created:       now,
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return "", errors.Wrapf(
			err,
			"error storing new root session %q",
			session.ID,
		)
	}
	return token, nil
}

func (s *service) CreateUserSession(
	ctx context.Context,
) (string, string, error) {
	oauth2State := crypto.NewToken(30)
	token := crypto.NewToken(256)
	session := brignext.Session{
		ID:                uuid.NewV4().String(),
		HashedOAuth2State: crypto.ShortSHA("", oauth2State),
		HashedToken:       crypto.ShortSHA("", token),
		Created:           time.Now(),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return "", "", errors.Wrapf(
			err,
			"error storing new user session %q",
			session.ID,
		)
	}
	return oauth2State, token, nil
}

func (s *service) GetSessionByOAuth2State(
	ctx context.Context,
	oauth2State string,
) (brignext.Session, error) {
	session, err := s.store.GetSessionByHashedOAuth2State(
		ctx,
		crypto.ShortSHA("", oauth2State),
	)
	if err != nil {
		return session, errors.Wrap(
			err,
			"error retrieving session from store by hashed oauth2 state",
		)
	}
	return session, nil
}

func (s *service) GetSessionByToken(
	ctx context.Context,
	token string,
) (brignext.Session, error) {
	session, err := s.store.GetSessionByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return session, errors.Wrap(
			err,
			"error retrieving session from store by hashed token",
		)
	}
	return session, nil
}

func (s *service) AuthenticateSession(
	ctx context.Context,
	sessionID string,
	userID string,
) error {
	if err := s.store.AuthenticateSession(
		ctx,
		sessionID,
		userID,
		time.Now().Add(time.Hour),
	); err != nil {
		return errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			sessionID,
		)
	}
	return nil
}

func (s *service) DeleteSession(ctx context.Context, id string) error {
	if err := s.store.DeleteSession(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

func (s *service) DeleteSessionsByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	n, err := s.store.DeleteSessionsByUser(ctx, userID)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error removing sessions for user %q from store",
			userID,
		)
	}
	return n, nil
}

func (s *service) CreateServiceAccount(
	ctx context.Context,
	serviceAccount brignext.ServiceAccount,
) (string, error) {
	if _, err := s.store.GetServiceAccount(ctx, serviceAccount.ID); err != nil {
		if _, ok := err.(*brignext.ErrServiceAccountNotFound); !ok {
			return "", errors.Wrapf(
				err,
				"error checking for existing service account %q",
				serviceAccount.ID,
			)
		}
	} else {
		return "", &brignext.ErrServiceAccountIDConflict{serviceAccount.ID}
	}

	token := crypto.NewToken(256)
	serviceAccount.HashedToken = crypto.ShortSHA("", token)
	now := time.Now()
	serviceAccount.Created = &now
	if err := s.store.CreateServiceAccount(ctx, serviceAccount); err != nil {
		return "", errors.Wrapf(
			err,
			"error storing new service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (s *service) GetServiceAccounts(
	ctx context.Context,
) ([]brignext.ServiceAccount, error) {
	serviceAccounts, err := s.store.GetServiceAccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccounts, nil
}

func (s *service) GetServiceAccount(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.GetServiceAccount(ctx, id)
	if err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *service) GetServiceAccountByToken(
	ctx context.Context,
	token string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.GetServiceAccountByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return serviceAccount, errors.Wrap(
			err,
			"error retrieving service account from store by hashed token",
		)
	}
	return serviceAccount, nil
}

func (s *service) LockServiceAccount(ctx context.Context, id string) error {
	if err := s.store.LockServiceAccount(ctx, id); err != nil {
		return errors.Wrapf(
			err,
			"error locking service account %q in the store",
			id,
		)
	}
	return nil
}

func (s *service) UnlockServiceAccount(
	ctx context.Context,
	id string,
) (string, error) {
	newToken := crypto.NewToken(256)
	if err := s.store.UnlockServiceAccount(
		ctx,
		id,
		crypto.ShortSHA("", newToken),
	); err != nil {
		return "", errors.Wrapf(
			err,
			"error unlocking service account %q in the store",
			id,
		)
	}
	return newToken, nil
}

func (s *service) CreateProject(
	ctx context.Context,
	project brignext.Project,
) error {
	if _, err := s.store.GetProject(ctx, project.ID); err != nil {
		if _, ok := err.(*brignext.ErrProjectNotFound); !ok {
			return errors.Wrapf(
				err,
				"error checking for existing project %q",
				project.ID,
			)
		}
	} else {
		return &brignext.ErrProjectIDConflict{project.ID}
	}

	namespace, err := s.scheduler.CreateProjectNamespace(project.ID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating new namespace for project %q",
			project.ID,
		)
	}

	project.Kubernetes = &brignext.ProjectKubernetesConfig{
		Namespace: namespace,
	}
	now := time.Now()
	project.Created = &now

	return s.store.DoTx(ctx, func(ctx context.Context) error {

		if err := s.store.CreateProject(ctx, project); err != nil {
			return errors.Wrapf(err, "error storing new project %q", project.ID)
		}

		secrets := project.Secrets
		if secrets == nil {
			secrets = map[string]string{}
		}

		if err := s.secretStore.CreateProjectSecrets(
			project.Kubernetes.Namespace,
			project.ID,
			secrets,
		); err != nil {
			return errors.Wrapf(err, "error storing project %q secrets", project.ID)
		}

		return nil
	})
}

func (s *service) GetProjects(ctx context.Context) ([]brignext.Project, error) {
	projects, err := s.store.GetProjects(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving projects from store")
	}
	return projects, nil
}

func (s *service) GetProject(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project, err := s.store.GetProject(ctx, id)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, nil
}

func (s *service) UpdateProject(
	ctx context.Context,
	project brignext.Project,
) error {
	return s.store.DoTx(ctx, func(ctx context.Context) error {

		if err := s.store.UpdateProject(ctx, project); err != nil {
			return errors.Wrapf(
				err,
				"error updating project %q in store",
				project.ID,
			)
		}

		secrets := project.Secrets
		if secrets == nil {
			secrets = map[string]string{}
		}

		// Get the updated project because it will contain a value in the namespace
		// field that the input didn't.
		project, err := s.store.GetProject(ctx, project.ID)
		if err != nil {
			return errors.Wrapf(
				err,
				"error retrieving updated project %q from store",
				project.ID,
			)
		}

		if err := s.secretStore.UpdateProjectSecrets(
			project.Kubernetes.Namespace,
			project.ID,
			secrets,
		); err != nil {
			return errors.Wrapf(err, "error updating project %q secrets", project.ID)
		}

		return nil
	})
}

func (s *service) DeleteProject(ctx context.Context, id string) error {
	project, err := s.store.GetProject(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	return s.store.DoTx(ctx, func(ctx context.Context) error {

		if err := s.store.DeleteProject(ctx, id); err != nil {
			return errors.Wrapf(err, "error removing project %q from store", id)
		}

		if err := s.store.DeleteEventsByProject(ctx, id); err != nil {
			return errors.Wrapf(
				err,
				"error removing events for project %q from store",
				id,
			)
		}

		// Deleting the namespace should take care of config maps, secrets, and
		// running pods as well.
		if err := s.scheduler.DeleteProjectNamespace(
			project.Kubernetes.Namespace,
		); err != nil {
			return errors.Wrapf(
				err,
				"error deleting namespace %q for project %q",
				project.Kubernetes.Namespace,
				id,
			)
		}

		return nil
	})
}

func (s *service) CreateEvent(
	ctx context.Context,
	event brignext.Event,
) (string, error) {
	// Make sure the project exists
	project, err := s.store.GetProject(ctx, event.ProjectID)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	event.ID = uuid.NewV4().String()
	event.Kubernetes = &brignext.EventKubernetesConfig{
		Namespace: project.Kubernetes.Namespace,
	}
	// "Split" the event into many workers.
	event.Workers = project.GetWorkers(event)
	if len(event.Workers) == 0 {
		event.Status = brignext.EventStatusMoot
	} else {
		event.Status = brignext.EventStatusPending
	}
	now := time.Now()
	event.Created = &now

	if err := s.store.DoTx(ctx, func(ctx context.Context) error {

		if err := s.store.CreateEvent(ctx, event); err != nil {
			return errors.Wrapf(err, "error storing new event %q", event.ID)
		}

		// This config map will contain a JSON file with event details. It will
		// get mounted into each of the event's worker pods.
		if err := s.secretStore.CreateEventConfigMap(event); err != nil {
			return errors.Wrapf(
				err,
				"error creating config map for event %q",
				event.ID,
			)
		}

		// This secret will be a snapshot of the project's secrets as they are in
		// this moment. Since all other aspects of the event and its workers are
		// based on a snapshot of the project configuration at the moment the event
		// was created, we need to do the same for the project's secrets. This way
		// project secrets are free to change, but the event will still be processed
		// using secrets as they were when the event was created.
		if err := s.secretStore.CreateEventSecrets(
			event.Kubernetes.Namespace,
			event.ProjectID,
			event.ID,
		); err != nil {
			return errors.Wrap(err, "error creating event secrets")
		}

		// This config maps will each contain a JSON file with worker details. Each
		// will get mounted into the applicable worker pod.
		for workerName, worker := range event.Workers {
			if err := s.secretStore.CreateWorkerConfigMap(
				event.Kubernetes.Namespace,
				event.ProjectID,
				event.ID,
				workerName,
				worker,
			); err != nil {
				return errors.Wrapf(
					err,
					"error creating config map for worker %q of new event %q",
					workerName,
					event.ID,
				)
			}
			if err := s.scheduler.ScheduleWorker(
				event.ProjectID,
				event.ID,
				workerName,
			); err != nil {
				return errors.Wrapf(
					err,
					"error scheduling worker %q for new event %q",
					workerName,
					event.ID,
				)
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	return event.ID, nil
}

func (s *service) GetEvents(ctx context.Context) ([]brignext.Event, error) {
	events, err := s.store.GetEvents(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving events from store")
	}
	return events, nil
}

func (s *service) GetEventsByProject(
	ctx context.Context,
	projectID string,
) ([]brignext.Event, error) {
	if _, err := s.store.GetProject(ctx, projectID); err != nil {
		return nil, errors.Wrapf(err, "error retrieving project %q", projectID)
	}
	events, err := s.store.GetEventsByProject(ctx, projectID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving events for project %q from store",
			projectID,
		)
	}
	return events, nil
}

func (s *service) GetEvent(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}
	return event, nil
}

func (s *service) UpdateEventStatus(
	ctx context.Context,
	id string,
	status brignext.EventStatus,
) error {
	if err := s.store.UpdateEventStatus(ctx, id, status); err != nil {
		return errors.Wrapf(err, "error updating event %q status in store", id)
	}
	return nil
}

func (s *service) UpdateEventWorkerStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	status brignext.WorkerStatus,
) error {
	if err := s.store.UpdateEventWorkerStatus(
		ctx,
		eventID,
		workerName,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status on worker %q of event %q in store",
			workerName,
			eventID,
		)
	}
	return nil
}

func (s *service) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteProcessing bool,
) (bool, error) {
	event, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	var ok bool
	err = s.store.DoTx(ctx, func(ctx context.Context) error {

		if ok, err = s.store.DeleteEvent(
			ctx,
			id,
			deletePending,
			deleteProcessing,
		); err != nil {
			return errors.Wrapf(err, "error removing event %q from store", id)
		}

		if ok {
			if err := s.scheduler.AbortWorkersByEvent(
				event.Kubernetes.Namespace,
				id,
			); err != nil {
				return errors.Wrapf(
					err,
					"error aborting running workers for event %q",
					id,
				)
			}

			if err := s.secretStore.DeleteEventConfigMaps(
				event.Kubernetes.Namespace,
				event.ID,
			); err != nil {
				return errors.Wrapf(
					err,
					"error deleting config maps for event %q",
					event.ID,
				)
			}

			if err := s.secretStore.DeleteEventSecrets(
				event.Kubernetes.Namespace,
				event.ID,
			); err != nil {
				return errors.Wrapf(
					err,
					"error deleting secrets for event %q",
					event.ID,
				)
			}
		}

		return nil
	})

	return ok, err
}

func (s *service) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
	deletePending bool,
	deleteProcessing bool,
) (int64, error) {

	// Find all events. We'll iterate over all of them and try to delete each.
	// It sounds inefficient and it probably is, but this allows us to delete
	// each event in its own transaction.
	events, err := s.store.GetEventsByProject(ctx, projectID)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error retrieving events for project %q",
			projectID,
		)
	}

	var deletedCount int64
	for _, event := range events {
		ok, err := s.DeleteEvent(ctx, event.ID, deletePending, deleteProcessing)
		if err != nil {
			return 0, errors.Wrapf(
				err,
				"error removing event %q from store",
				event.ID,
			)
		}
		if ok {
			deletedCount++
		}
	}

	return deletedCount, nil
}
