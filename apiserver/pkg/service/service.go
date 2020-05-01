package service

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/crypto"
	"github.com/krancour/brignext/apiserver/pkg/scheduler"
	"github.com/krancour/brignext/apiserver/pkg/storage"
	"github.com/krancour/brignext/pkg/retries"
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
		workerName string,
	) (map[string]string, error)
	SetSecrets(
		ctx context.Context,
		projectID string,
		workerName string,
		secrets map[string]string,
	) error
	UnsetSecrets(
		ctx context.Context,
		projectID string,
		workerName string,
		keys []string,
	) error

	CreateEvent(context.Context, brignext.Event) (string, error)
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, error)
	CancelEvent(
		ctx context.Context,
		id string,
		cancelProcessing bool,
	) (bool, error)
	CancelEventsByProject(
		ctx context.Context,
		projectID string,
		cancelingProcessing bool,
	) (int64, error)
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

	GetWorker(
		ctx context.Context,
		eventID string,
		workerName string,
	) (brignext.Worker, error)
	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		status brignext.WorkerStatus,
	) error
	CancelWorker(
		ctx context.Context,
		eventID string,
		workerName string,
		cancelRunning bool,
	) (bool, error)
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]brignext.LogEntry, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]brignext.LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan brignext.LogEntry, error)

	GetJob(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (brignext.Job, error)
	UpdateJobStatus(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
		status brignext.JobStatus,
	) error
	GetJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
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
	for workerName, workerConfig := range project.WorkerConfigs {
		if workerConfig.LogLevel == "" {
			workerConfig.LogLevel = brignext.LogLevelInfo
		}
		project.WorkerConfigs[workerName] = workerConfig
	}
	now := time.Now()
	project.Created = &now
	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transation
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
	// This is in contrast to most of our functions wherein we start a transation
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
	workerName string,
) (map[string]string, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if _, ok := project.WorkerConfigs[workerName]; !ok {
		return nil, &brignext.ErrWorkerNotFound{
			ProjectID:  projectID,
			WorkerName: workerName,
		}
	}
	secrets, err := s.scheduler.GetSecrets(ctx, project, workerName)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error getting secrets for project %q worker %q from scheduler",
			projectID,
			workerName,
		)
	}
	return secrets, nil
}

func (s *service) SetSecrets(
	ctx context.Context,
	projectID string,
	workerName string,
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
	if _, ok := project.WorkerConfigs[workerName]; !ok {
		return &brignext.ErrWorkerNotFound{
			ProjectID:  projectID,
			WorkerName: workerName,
		}
	}
	// Secrets aren't stored in the database. We only pass them to the scheduler.
	if err := s.scheduler.SetSecrets(
		ctx,
		project,
		workerName,
		secrets,
	); err != nil {
		return errors.Wrapf(
			err,
			"error setting secrets for project %q worker %q in scheduler",
			projectID,
			workerName,
		)
	}
	return nil
}

func (s *service) UnsetSecrets(
	ctx context.Context,
	projectID string,
	workerName string,
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
	if _, ok := project.WorkerConfigs[workerName]; !ok {
		return &brignext.ErrWorkerNotFound{
			ProjectID:  projectID,
			WorkerName: workerName,
		}
	}
	// Secrets aren't stored in the database. We only have to remove them from the
	// scheduler.
	if err := s.scheduler.UnsetSecrets(
		ctx,
		project,
		workerName,
		keys,
	); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker %q in scheduler",
			projectID,
			workerName,
		)
	}
	return nil
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
	// "Split" the event into many workers.
	event.Workers = project.GetWorkers(event)
	now := time.Now()
	event.Created = &now
	event.Status = &brignext.EventStatus{}
	if len(event.Workers) == 0 {
		event.Status.Phase = brignext.EventPhaseMoot
	} else {
		event.Status.Phase = brignext.EventPhasePending
	}

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transation
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	event, err = s.scheduler.CreateEvent(
		ctx,
		project,
		event,
	)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}
	if err := s.store.CreateEvent(ctx, event); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.DeleteEvent(ctx, event) // nolint: errcheck
		return "", errors.Wrapf(err, "error storing new event %q", event.ID)
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
	cancelProcessing bool,
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
			cancelProcessing,
		); err != nil {
			return errors.Wrapf(err, "error updating event %q in store", id)
		}
		if ok {
			if err := s.scheduler.DeleteEvent(ctx, event); err != nil {
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

func (s *service) CancelEventsByProject(
	ctx context.Context,
	projectID string,
	cancelProcessing bool,
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
			ok, err := s.store.CancelEvent(ctx, event.ID, cancelProcessing)
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
			if err := s.scheduler.DeleteEvent(ctx, event); err != nil {
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
		if err := s.store.DoTx(ctx, func(ctx context.Context) error {
			ok, err := s.store.DeleteEvent(
				ctx,
				event.ID,
				deletePending,
				deleteProcessing,
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

func (s *service) GetWorker(
	ctx context.Context,
	eventID string,
	workerName string,
) (brignext.Worker, error) {
	worker, err := s.store.GetWorker(ctx, eventID, workerName)
	if err != nil {
		return worker, errors.Wrapf(
			err,
			"error retrieving event %q worker %q from store",
			eventID,
			workerName,
		)
	}
	return worker, nil
}

// TODO: This logic isn't correct!!! It must be fixed!!!
func (s *service) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	status brignext.WorkerStatus,
) error {
	if err := s.store.UpdateWorkerStatus(
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

	retries.ManageRetries(
		ctx,
		fmt.Sprintf("obtaining lock on event %q", eventID),
		5,
		10*time.Second,
		func() (bool, error) {
			locked, err := s.store.LockEvent(ctx, eventID)
			if err != nil {
				// s.store.LockEvent(...) reports success or failure in obtaining a lock
				// using a boolean. Any error wasn't merely inability to obtain a lock.
				// It was something else unexpected and we should not retry.
				return false, err
			}
			// Retry is indicated if a lock was NOT obtained.
			return !locked, nil
		},
	)

	event, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	// Bail early if the event is already in a terminal phase
	switch event.Status.Phase {
	case brignext.EventPhasePending:
	case brignext.EventPhaseProcessing:
	default:
		return nil
	}

	// Determine how collective worker phases change event phase...

	var anyStarted bool // Will be true if any worker has (at least) started
	var anyFailed bool  // Will be true if any worker has failed
	// Will remain true only if all workers are in a terminal state
	allFinished := true

	eventStatus := brignext.EventStatus{}
	var latestWorkerEnd *time.Time

	for _, worker := range event.Workers {
		if eventStatus.Started == nil ||
			(worker.Status.Started != nil &&
				worker.Status.Started.Before(*eventStatus.Started)) {
			eventStatus.Started = worker.Status.Started
		}
		if latestWorkerEnd == nil ||
			(worker.Status.Ended != nil &&
				worker.Status.Ended.After(*latestWorkerEnd)) {
			latestWorkerEnd = worker.Status.Ended
		}
		switch worker.Status.Phase {
		case brignext.WorkerPhasePending:
			allFinished = false
		case brignext.WorkerPhaseRunning:
			anyStarted = true
			allFinished = false
		case brignext.WorkerPhaseSucceeded:
			anyStarted = true
		case brignext.WorkerPhaseFailed:
			anyStarted = true
			anyFailed = true
		}
	}

	// Note there are no transitions to aborted or canceled phases here. Those are
	// handled separately, from the top down, instead of determining event phase
	// based on worker phases like we are doing here.
	if !anyStarted {
		eventStatus.Phase = brignext.EventPhasePending
	} else if allFinished {
		// We're done-- figure out more specific state
		if anyFailed {
			eventStatus.Phase = brignext.EventPhaseFailed
		} else {
			eventStatus.Phase = brignext.EventPhaseSucceeded
		}
	} else {
		// We've started, but haven't finished, so we're processing
		eventStatus.Phase = brignext.EventPhaseProcessing
	}

	if allFinished {
		eventStatus.Ended = latestWorkerEnd
	}

	if err := s.store.UpdateEventStatus(
		ctx,
		eventID,
		eventStatus,
	); err != nil {
		return errors.Wrapf(err, "error updating event %q status in store", eventID)
	}

	if err := s.store.UnlockEvent(ctx, eventID); err != nil {
		return errors.Wrapf(err, "error releasing lock on event %q", eventID)
	}

	return nil
}

func (s *service) CancelWorker(
	ctx context.Context,
	eventID string,
	workerName string,
	cancelRunning bool,
) (bool, error) {
	event, err := s.store.GetEvent(ctx, eventID)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			eventID,
		)
	}

	worker, ok := event.Workers[workerName]
	if !ok {
		return false, &brignext.ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}

	if worker.Status.Phase == brignext.WorkerPhasePending {
		worker.Status.Phase = brignext.WorkerPhaseCanceled
	} else if cancelRunning &&
		worker.Status.Phase == brignext.WorkerPhaseRunning {
		worker.Status.Phase = brignext.WorkerPhaseAborted
	} else {
		return false, nil
	}

	err = s.store.DoTx(ctx, func(ctx context.Context) error {
		if err = s.store.UpdateWorkerStatus(
			ctx,
			eventID,
			workerName,
			worker.Status,
		); err != nil {
			return errors.Wrapf(
				err,
				"error updating event %q worker %q status in store",
				eventID,
				workerName,
			)
		}
		if err := s.scheduler.DeleteWorker(ctx, event, workerName); err != nil {
			return errors.Wrapf(
				err,
				"error deleting event %q worker %q from scheduler",
				eventID,
				workerName,
			)
		}
		return nil
	})

	return true, nil
}

func (s *service) GetWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetWorkerLogs(ctx, eventID, workerName)
}

func (s *service) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamWorkerLogs(ctx, eventID, workerName)
}

func (s *service) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetWorkerInitLogs(ctx, eventID, workerName)
}

func (s *service) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamWorkerInitLogs(ctx, eventID, workerName)
}

func (s *service) GetJob(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (brignext.Job, error) {
	job, err := s.store.GetJob(ctx, eventID, workerName, jobName)
	if err != nil {
		return job, errors.Wrapf(
			err,
			"error retrieving event %q worker %q job %q from store",
			eventID,
			workerName,
			jobName,
		)
	}
	return job, nil
}

func (s *service) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
	status brignext.JobStatus,
) error {
	if err := s.store.UpdateJobStatus(
		ctx,
		eventID,
		workerName,
		jobName,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status on worker %q job %q of event %q in store",
			workerName,
			jobName,
			eventID,
		)
	}
	return nil
}

func (s *service) GetJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetJobLogs(ctx, eventID, workerName, jobName)
}

func (s *service) StreamJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamJobLogs(ctx, eventID, workerName, jobName)
}

func (s *service) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return s.logStore.GetJobInitLogs(ctx, eventID, workerName, jobName)
}

func (s *service) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamJobInitLogs(ctx, eventID, workerName, jobName)
}
