package service

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
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

	GetSecrets(
		ctx context.Context,
		projectID string,
	) (map[string]string, error)
	SetSecrets(
		ctx context.Context,
		projectID string,
		secrets map[string]string,
	) error
	UnsetSecrets(
		ctx context.Context,
		projectID string,
		keys []string,
	) error

	CreateEvent(context.Context, brignext.Event) ([]string, error)
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, error)
	CancelEvent(
		ctx context.Context,
		id string,
		cancelRunning bool,
	) (bool, error)
	CancelEventsByProject(
		ctx context.Context,
		projectID string,
		cancelingRunning bool,
	) (int64, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteRunning bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deletePending bool,
		deleteRunning bool,
	) (int64, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
	) ([]brignext.LogEntry, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) ([]brignext.LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)

	GetJob(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.Job, error)
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status brignext.JobStatus,
	) error
	GetJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
}

type service struct {
	store     storage.Store
	scheduler scheduler.Scheduler
	logStore  storage.LogStore
}

func NewService(
	store storage.Store,
	scheduler scheduler.Scheduler,
	logStore storage.LogStore,
) Service {
	return &service{
		store:     store,
		scheduler: scheduler,
		logStore:  logStore,
	}
}

func (s *service) CreateUser(ctx context.Context, user brignext.User) error {
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
		Expires:       now.Add(time.Hour),
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
	// TODO: Move where we set this to where we set the other defaults-- i.e.
	// set it at the time an event is created-- not when the project is created.
	if project.Spec.WorkerConfig.LogLevel == "" {
		project.Spec.WorkerConfig.LogLevel = brignext.LogLevelInfo
	}

	now := time.Now()
	project.Created = &now
	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transaction
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	var err error
	project, err = s.scheduler.CreateProject(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}
	if err := s.store.CreateProject(ctx, project); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.DeleteProject(ctx, project) // nolint: errcheck
		return errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	return nil
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
	// Get the original project in case we need to do a manual rollback
	oldProject, err := s.store.GetProject(ctx, project.ID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			project.ID,
		)
	}
	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transaction
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	if project, err = s.scheduler.UpdateProject(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in scheduler",
			project.ID,
		)
	}
	if err := s.store.UpdateProject(ctx, project); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.UpdateProject(ctx, oldProject) // nolint: errcheck
		return errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return nil
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
		if err := s.scheduler.DeleteProject(ctx, project); err != nil {
			return errors.Wrapf(
				err,
				"error deleting project %q from scheduler",
				id,
			)
		}
		return nil
	})
}

func (s *service) GetSecrets(
	ctx context.Context,
	projectID string,
) (map[string]string, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	secrets, err := s.scheduler.GetSecrets(ctx, project)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from scheduler",
			projectID,
		)
	}
	return secrets, nil
}

func (s *service) SetSecrets(
	ctx context.Context,
	projectID string,
	secrets map[string]string,
) error {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only pass them to the scheduler.
	if err := s.scheduler.SetSecrets(ctx, project, secrets); err != nil {
		return errors.Wrapf(
			err,
			"error setting secrets for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func (s *service) UnsetSecrets(
	ctx context.Context,
	projectID string,
	keys []string,
) error {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only have to remove them from the
	// scheduler.
	if err := s.scheduler.UnsetSecrets(ctx, project, keys); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func (s *service) CreateEvent(
	ctx context.Context,
	event brignext.Event,
) ([]string, error) {

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projects, err := s.store.GetSubscribedProjects(ctx, event)
		if err != nil {
			return nil, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		eventIDs := make([]string, len(projects))
		for i, project := range projects {
			event.ProjectID = project.ID
			eids, err := s.CreateEvent(ctx, event)
			if err != nil {
				return eventIDs, err
			}
			// eids will always contain precisely one element
			eventIDs[i] = eids[0]
		}
		return eventIDs, nil
	}

	// Make sure the project exists
	project, err := s.store.GetProject(ctx, event.ProjectID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	event.ID = uuid.NewV4().String()

	event.Worker = &brignext.Worker{
		Container:            project.Spec.WorkerConfig.Container,
		WorkspaceSize:        project.Spec.WorkerConfig.WorkspaceSize,
		Git:                  project.Spec.WorkerConfig.Git,
		JobsConfig:           project.Spec.WorkerConfig.JobsConfig,
		LogLevel:             project.Spec.WorkerConfig.LogLevel,
		ConfigFilesDirectory: project.Spec.WorkerConfig.ConfigFilesDirectory,
		DefaultConfigFiles:   project.Spec.WorkerConfig.DefaultConfigFiles,
		Jobs:                 map[string]brignext.Job{},
		Status: brignext.WorkerStatus{
			Phase: brignext.WorkerPhasePending,
		},
	}
	if event.Worker.WorkspaceSize == "" {
		event.Worker.WorkspaceSize = "10Gi"
	}

	// VCS details from the event override project-level details
	if event.Git.CloneURL != "" {
		event.Worker.Git.CloneURL = event.Git.CloneURL
	}
	if event.Git.Commit != "" {
		event.Worker.Git.Commit = event.Git.Commit
	}
	if event.Git.Ref != "" {
		event.Worker.Git.Ref = event.Git.Ref
	}

	if event.Worker.Git.CloneURL != "" &&
		event.Worker.Git.Commit == "" &&
		event.Worker.Git.Ref == "" {
		event.Worker.Git.Ref = "master"
	}

	if event.Worker.ConfigFilesDirectory == "" {
		event.Worker.ConfigFilesDirectory = "."
	}

	now := time.Now()
	event.Created = &now

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transaction
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	event, err = s.scheduler.CreateEvent(
		ctx,
		project,
		event,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}
	if err := s.store.CreateEvent(ctx, event); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.DeleteEvent(ctx, event) // nolint: errcheck
		return nil, errors.Wrapf(err, "error storing new event %q", event.ID)
	}
	return []string{event.ID}, nil
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
	if event, err = s.scheduler.GetEvent(ctx, event); err != nil {
		return event, errors.Wrapf(
			err,
			"error retrieving event %q from scheduler",
			id,
		)
	}
	return event, nil
}

func (s *service) CancelEvent(
	ctx context.Context,
	id string,
	cancelRunning bool,
) (bool, error) {
	event, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	var ok bool
	err = s.store.DoTx(ctx, func(ctx context.Context) error {
		if ok, err = s.store.CancelEvent(
			ctx,
			id,
			cancelRunning,
		); err != nil {
			return errors.Wrapf(err, "error updating event %q in store", id)
		}
		if ok {
			if err = s.scheduler.CancelEvent(ctx, event); err != nil {
				return errors.Wrapf(
					err,
					"error canceling event %q in scheduler",
					id,
				)
			}
		}
		return nil
	})

	return ok, err
}

func (s *service) CancelEventsByProject(
	ctx context.Context,
	projectID string,
	cancelRunning bool,
) (int64, error) {

	// Find all events. We'll iterate over all of them and try to cancel each.
	// It sounds inefficient and it probably is, but this allows us to cancel
	// each event in its own transaction.
	events, err := s.store.GetEventsByProject(ctx, projectID)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error retrieving events for project %q",
			projectID,
		)
	}

	var canceledCount int64
	for _, event := range events {
		if err := s.store.DoTx(ctx, func(ctx context.Context) error {
			ok, err := s.store.CancelEvent(ctx, event.ID, cancelRunning)
			if err != nil {
				return errors.Wrapf(
					err,
					"error canceling event %q in store",
					event.ID,
				)
			}
			if ok {
				canceledCount++
				if err := s.scheduler.DeleteEvent(ctx, event); err != nil {
					return errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					)
				}
			}
			return nil
		}); err != nil {
			return canceledCount, err
		}
	}

	return canceledCount, nil
}

func (s *service) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
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
			deleteRunning,
		); err != nil {
			return errors.Wrapf(err, "error removing event %q from store", id)
		}
		if ok {
			if err = s.scheduler.DeleteEvent(ctx, event); err != nil {
				return errors.Wrapf(
					err,
					"error deleting event %q from scheduler",
					id,
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
	deleteRunning bool,
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
		if err := s.store.DoTx(ctx, func(ctx context.Context) error {
			ok, err := s.store.DeleteEvent(
				ctx,
				event.ID,
				deletePending,
				deleteRunning,
			)
			if err != nil {
				return errors.Wrapf(
					err,
					"error deleting event %q from store",
					event.ID,
				)
			}
			if ok {
				deletedCount++
				if err := s.scheduler.DeleteEvent(ctx, event); err != nil {
					return errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					)
				}
			}
			return nil
		}); err != nil {
			return deletedCount, err
		}
	}

	return deletedCount, nil
}

func (s *service) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	if err := s.store.UpdateWorkerStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}
	return nil
}

func (s *service) GetWorkerLogs(
	ctx context.Context,
	eventID string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetWorkerLogs(ctx, eventID)
}

func (s *service) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamWorkerLogs(ctx, eventID)
}

func (s *service) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetWorkerInitLogs(ctx, eventID)
}

func (s *service) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamWorkerInitLogs(ctx, eventID)
}

func (s *service) GetJob(
	ctx context.Context,
	eventID string,
	jobName string,
) (brignext.Job, error) {
	job, err := s.store.GetJob(ctx, eventID, jobName)
	if err != nil {
		return job, errors.Wrapf(
			err,
			"error retrieving event %q worker job %q from store",
			eventID,
			jobName,
		)
	}
	return job, nil
}

func (s *service) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	if err := s.store.UpdateJobStatus(
		ctx,
		eventID,
		jobName,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q in store",
			eventID,
			jobName,
		)
	}
	return nil
}

func (s *service) GetJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetJobLogs(ctx, eventID, jobName)
}

func (s *service) StreamJobLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamJobLogs(ctx, eventID, jobName)
}

func (s *service) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetJobInitLogs(ctx, eventID, jobName)
}

func (s *service) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamJobInitLogs(ctx, eventID, jobName)
}
