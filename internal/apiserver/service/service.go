package service

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto"
	"github.com/krancour/brignext/v2/internal/apiserver/scheduler"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Service interface {
	CreateUser(context.Context, brignext.User) error
	GetUsers(context.Context) (brignext.UserList, error)
	GetUser(context.Context, string) (brignext.User, error)
	LockUser(context.Context, string) error
	UnlockUser(context.Context, string) error

	CreateRootSession(context.Context) (brignext.Token, error)
	CreateUserSession(context.Context) (string, string, error)
	GetSessionByOAuth2State(context.Context, string) (auth.Session, error)
	GetSessionByToken(context.Context, string) (auth.Session, error)
	AuthenticateSession(
		ctx context.Context,
		sessionID string,
		userID string,
	) error
	DeleteSession(context.Context, string) error
	DeleteSessionsByUser(context.Context, string) (int64, error)

	CreateServiceAccount(
		context.Context,
		brignext.ServiceAccount,
	) (brignext.Token, error)
	GetServiceAccounts(context.Context) (brignext.ServiceAccountList, error)
	GetServiceAccount(context.Context, string) (brignext.ServiceAccount, error)
	GetServiceAccountByToken(
		context.Context,
		string,
	) (brignext.ServiceAccount, error)
	LockServiceAccount(context.Context, string) error
	UnlockServiceAccount(context.Context, string) (brignext.Token, error)

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) (brignext.ProjectList, error)
	GetProject(context.Context, string) (brignext.Project, error)
	UpdateProject(context.Context, brignext.Project) error
	DeleteProject(context.Context, string) error

	GetSecrets(
		ctx context.Context,
		projectID string,
	) (brignext.SecretList, error)
	SetSecret(
		ctx context.Context,
		projectID string,
		secret brignext.Secret,
	) error
	UnsetSecret(
		ctx context.Context,
		projectID string,
		secretID string,
	) error

	CreateEvent(context.Context, brignext.Event) (
		brignext.EventReferenceList,
		error,
	)
	GetEvents(context.Context) (brignext.EventList, error)
	GetEventsByProject(context.Context, string) (brignext.EventList, error)
	GetEvent(context.Context, string) (brignext.Event, error)
	CancelEvent(
		ctx context.Context,
		id string,
		cancelRunning bool,
	) (brignext.EventReferenceList, error)
	CancelEventsByProject(
		ctx context.Context,
		projectID string,
		cancelingRunning bool,
	) (brignext.EventReferenceList, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deletePending bool,
		deleteRunning bool,
	) (brignext.EventReferenceList, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deletePending bool,
		deleteRunning bool,
	) (brignext.EventReferenceList, error)

	UpdateWorkerStatus(
		ctx context.Context,
		eventID string,
		status brignext.WorkerStatus,
	) error
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)

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
	) (brignext.LogEntryList, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.LogEntryList, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
}

type service struct {
	store     storage.Store
	scheduler scheduler.Scheduler
	logStore  storage.LogsStore
}

func NewService(
	store storage.Store,
	scheduler scheduler.Scheduler,
	logStore storage.LogsStore,
) Service {
	return &service{
		store:     store,
		scheduler: scheduler,
		logStore:  logStore,
	}
}

func (s *service) CreateUser(ctx context.Context, user brignext.User) error {
	if err := s.store.Users().Create(ctx, user); err != nil {
		return errors.Wrapf(err, "error storing new user %q", user.ID)
	}
	return nil
}

func (s *service) GetUsers(ctx context.Context) (brignext.UserList, error) {
	userList, err := s.store.Users().List(ctx)
	if err != nil {
		return userList, errors.Wrap(err, "error retrieving users from store")
	}
	return userList, nil
}

func (s *service) GetUser(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user, err := s.store.Users().Get(ctx, id)
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
		if err = s.store.Users().Lock(ctx, id); err != nil {
			return errors.Wrapf(err, "error locking user %q in store", id)
		}
		if _, err := s.store.Sessions().DeleteByUser(ctx, id); err != nil {
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
	if err := s.store.Users().Unlock(ctx, id); err != nil {
		return errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return nil
}

func (s *service) CreateRootSession(
	ctx context.Context,
) (brignext.Token, error) {
	token := brignext.Token{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Token",
		},
		Value: crypto.NewToken(256),
	}
	now := time.Now()
	expiryTime := now.Add(time.Hour)
	session := auth.Session{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: brignext.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		Root:          true,
		HashedToken:   crypto.ShortSHA("", token.Value),
		Authenticated: &now,
		Expires:       &expiryTime,
	}
	if err := s.store.Sessions().Create(ctx, session); err != nil {
		return token, errors.Wrapf(
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
	session := auth.Session{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Session",
		},
		ObjectMeta: brignext.ObjectMeta{
			ID: uuid.NewV4().String(),
		},
		HashedOAuth2State: crypto.ShortSHA("", oauth2State),
		HashedToken:       crypto.ShortSHA("", token),
	}
	if err := s.store.Sessions().Create(ctx, session); err != nil {
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
) (auth.Session, error) {
	session, err := s.store.Sessions().GetByHashedOAuth2State(
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
) (auth.Session, error) {
	session, err := s.store.Sessions().GetByHashedToken(
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
	if err := s.store.Sessions().Authenticate(
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
	if err := s.store.Sessions().Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing session %q from store", id)
	}
	return nil
}

func (s *service) DeleteSessionsByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	n, err := s.store.Sessions().DeleteByUser(ctx, userID)
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
) (brignext.Token, error) {
	token := brignext.Token{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Token",
		},
		Value: crypto.NewToken(256),
	}
	serviceAccount.HashedToken = crypto.ShortSHA("", token.Value)
	if err := s.store.ServiceAccounts().Create(ctx, serviceAccount); err != nil {
		return token, errors.Wrapf(
			err,
			"error storing new service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (s *service) GetServiceAccounts(
	ctx context.Context,
) (brignext.ServiceAccountList, error) {
	serviceAccountList, err := s.store.ServiceAccounts().List(ctx)
	if err != nil {
		return serviceAccountList,
			errors.Wrap(err, "error retrieving service accounts from store")
	}
	return serviceAccountList, nil
}

func (s *service) GetServiceAccount(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount, err := s.store.ServiceAccounts().Get(ctx, id)
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
	serviceAccount, err := s.store.ServiceAccounts().GetByHashedToken(
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
	if err := s.store.ServiceAccounts().Lock(ctx, id); err != nil {
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
) (brignext.Token, error) {
	newToken := brignext.Token{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "Token",
		},
		Value: crypto.NewToken(256),
	}
	if err := s.store.ServiceAccounts().Unlock(
		ctx,
		id,
		crypto.ShortSHA("", newToken.Value),
	); err != nil {
		return newToken, errors.Wrapf(
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
	if project.Spec.Worker.LogLevel == "" {
		project.Spec.Worker.LogLevel = brignext.LogLevelInfo
	}

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transaction
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	var err error
	project, err = s.scheduler.Projects().Create(ctx, project)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating project %q in the scheduler",
			project.ID,
		)
	}
	if err := s.store.Projects().Create(ctx, project); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.Projects().Delete(ctx, project) // nolint: errcheck
		return errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	return nil
}

func (s *service) GetProjects(
	ctx context.Context,
) (brignext.ProjectList, error) {
	projectList, err := s.store.Projects().List(ctx)
	if err != nil {
		return projectList, errors.Wrap(err, "error retrieving projects from store")
	}
	return projectList, nil
}

func (s *service) GetProject(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project, err := s.store.Projects().Get(ctx, id)
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
	if err := s.store.Projects().Update(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return nil
}

func (s *service) DeleteProject(ctx context.Context, id string) error {
	project, err := s.store.Projects().Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}
	return s.store.DoTx(ctx, func(ctx context.Context) error {
		if err := s.store.Projects().Delete(ctx, id); err != nil {
			return errors.Wrapf(err, "error removing project %q from store", id)
		}
		if err := s.store.Events().DeleteByProject(ctx, id); err != nil {
			return errors.Wrapf(
				err,
				"error deleting events for project %q from scheduler",
				id,
			)
		}
		if err := s.scheduler.Projects().Delete(ctx, project); err != nil {
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
) (brignext.SecretList, error) {
	secretList := brignext.SecretList{}
	project, err := s.store.Projects().Get(ctx, projectID)
	if err != nil {
		return secretList, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if secretList, err =
		s.scheduler.Projects().ListSecrets(ctx, project); err != nil {
		return secretList, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from scheduler",
			projectID,
		)
	}
	return secretList, nil
}

func (s *service) SetSecret(
	ctx context.Context,
	projectID string,
	secret brignext.Secret,
) error {
	project, err := s.store.Projects().Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only pass them to the scheduler.
	if err := s.scheduler.Projects().SetSecret(ctx, project, secret); err != nil {
		return errors.Wrapf(
			err,
			"error setting secret for project %q worker in scheduler",
			projectID,
		)
	}
	return nil
}

func (s *service) UnsetSecret(
	ctx context.Context,
	projectID string,
	secretID string,
) error {
	project, err := s.store.Projects().Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	// Secrets aren't stored in the database. We only have to remove them from the
	// scheduler.
	if err :=
		s.scheduler.Projects().UnsetSecret(ctx, project, secretID); err != nil {
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
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: brignext.ListMeta{},
	}

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projectList, err := s.store.Projects().ListSubscribed(ctx, event)
		if err != nil {
			return eventRefList, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		eventRefList.Items = make([]brignext.EventReference, len(projectList.Items))
		for i, project := range projectList.Items {
			event.ProjectID = project.ID
			eRefs, err := s.CreateEvent(ctx, event)
			if err != nil {
				return eventRefList, err
			}
			// eids will always contain precisely one element
			eventRefList.Items[i] = eRefs.Items[0]
		}
		return eventRefList, nil
	}

	// Make sure the project exists
	project, err := s.store.Projects().Get(ctx, event.ProjectID)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	event.ID = uuid.NewV4().String()

	event.Worker = &project.Spec.Worker

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

	event.Status = &brignext.EventStatus{
		WorkerStatus: brignext.WorkerStatus{
			Phase: brignext.WorkerPhasePending,
		},
		JobStatuses: map[string]brignext.JobStatus{},
	}

	// We send this to the scheduler first because we expect the scheduler will
	// will add some scheduler-specific details that we will want to persist.
	// This is in contrast to most of our functions wherein we start a transaction
	// in the store and make modifications to the store first with expectations
	// that the transaction will roll the change back if subsequent changes made
	// via the scheduler fail.
	event, err = s.scheduler.Events().Create(
		ctx,
		project,
		event,
	)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error creating event %q in scheduler",
			event.ID,
		)
	}
	if err := s.store.Events().Create(ctx, event); err != nil {
		// We need to roll this back manually because the scheduler doesn't
		// automatically roll anything back upon failure.
		s.scheduler.Events().Delete(ctx, event) // nolint: errcheck
		return eventRefList,
			errors.Wrapf(err, "error storing new event %q", event.ID)
	}

	eventRefList.Items = []brignext.EventReference{
		{
			TypeMeta: brignext.TypeMeta{
				APIVersion: brignext.APIVersion,
				Kind:       "EventReference",
			},
			ID: event.ID,
		},
	}
	return eventRefList, nil
}

func (s *service) GetEvents(ctx context.Context) (brignext.EventList, error) {
	eventList, err := s.store.Events().List(ctx)
	if err != nil {
		return eventList, errors.Wrap(err, "error retrieving events from store")
	}
	return eventList, nil
}

func (s *service) GetEventsByProject(
	ctx context.Context,
	projectID string,
) (brignext.EventList, error) {
	if _, err := s.store.Projects().Get(ctx, projectID); err != nil {
		return brignext.EventList{},
			errors.Wrapf(err, "error retrieving project %q", projectID)
	}
	eventList, err := s.store.Events().ListByProject(ctx, projectID)
	if err != nil {
		return eventList, errors.Wrapf(
			err,
			"error retrieving events for project %q from store",
			projectID,
		)
	}
	return eventList, nil
}

func (s *service) GetEvent(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event, err := s.store.Events().Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}
	if event, err = s.scheduler.Events().Get(ctx, event); err != nil {
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
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: brignext.ListMeta{},
		Items:    []brignext.EventReference{},
	}

	event, err := s.store.Events().Get(ctx, id)
	if err != nil {
		return eventRefList,
			errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	err = s.store.DoTx(ctx, func(ctx context.Context) error {
		var ok bool
		ok, err = s.store.Events().Cancel(
			ctx,
			id,
			cancelRunning,
		)
		if err != nil {
			return errors.Wrapf(err, "error updating event %q in store", id)
		}
		if ok {
			if err = s.scheduler.Events().Cancel(ctx, event); err != nil {
				return errors.Wrapf(
					err,
					"error canceling event %q in scheduler",
					id,
				)
			}
			eventRefList.Items = append(
				eventRefList.Items,
				brignext.EventReference{
					TypeMeta: brignext.TypeMeta{
						APIVersion: brignext.APIVersion,
						Kind:       "EventReference",
					},
					ID: event.ID,
				},
			)
		}
		return nil
	})

	return eventRefList, err
}

func (s *service) CancelEventsByProject(
	ctx context.Context,
	projectID string,
	cancelRunning bool,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: brignext.ListMeta{},
		Items:    []brignext.EventReference{},
	}

	// Find all events. We'll iterate over all of them and try to cancel each.
	// It sounds inefficient and it probably is, but this allows us to cancel
	// each event in its own transaction.
	eventList, err := s.store.Events().ListByProject(ctx, projectID)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error retrieving events for project %q",
			projectID,
		)
	}

	for _, event := range eventList.Items {
		if err := s.store.DoTx(ctx, func(ctx context.Context) error {
			ok, err := s.store.Events().Cancel(ctx, event.ID, cancelRunning)
			if err != nil {
				return errors.Wrapf(
					err,
					"error canceling event %q in store",
					event.ID,
				)
			}
			if ok {
				if err := s.scheduler.Events().Delete(ctx, event); err != nil {
					return errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					)
				}
				eventRefList.Items = append(
					eventRefList.Items,
					brignext.EventReference{
						TypeMeta: brignext.TypeMeta{
							APIVersion: brignext.APIVersion,
							Kind:       "EventReference",
						},
						ID: event.ID,
					},
				)
			}
			return nil
		}); err != nil {
			return eventRefList, err
		}
	}

	return eventRefList, nil
}

func (s *service) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: brignext.ListMeta{},
		Items:    []brignext.EventReference{},
	}

	event, err := s.store.Events().Get(ctx, id)
	if err != nil {
		return eventRefList,
			errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	err = s.store.DoTx(ctx, func(ctx context.Context) error {
		var ok bool
		ok, err = s.store.Events().Delete(
			ctx,
			id,
			deletePending,
			deleteRunning,
		)
		if err != nil {
			return errors.Wrapf(err, "error removing event %q from store", id)
		}
		if ok {
			if err = s.scheduler.Events().Delete(ctx, event); err != nil {
				return errors.Wrapf(
					err,
					"error deleting event %q from scheduler",
					id,
				)
			}
			eventRefList.Items = append(
				eventRefList.Items,
				brignext.EventReference{
					TypeMeta: brignext.TypeMeta{
						APIVersion: brignext.APIVersion,
						Kind:       "EventReference",
					},
					ID: event.ID,
				},
			)
		}
		return nil
	})
	return eventRefList, err
}

func (s *service) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
	deletePending bool,
	deleteRunning bool,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.EventReferenceList{
		TypeMeta: brignext.TypeMeta{
			APIVersion: brignext.APIVersion,
			Kind:       "EventReferenceList",
		},
		ListMeta: brignext.ListMeta{},
		Items:    []brignext.EventReference{},
	}

	// Find all events. We'll iterate over all of them and try to delete each.
	// It sounds inefficient and it probably is, but this allows us to delete
	// each event in its own transaction.
	eventList, err := s.store.Events().ListByProject(ctx, projectID)
	if err != nil {
		return eventRefList, errors.Wrapf(
			err,
			"error retrieving events for project %q",
			projectID,
		)
	}

	for _, event := range eventList.Items {
		if err := s.store.DoTx(ctx, func(ctx context.Context) error {
			ok, err := s.store.Events().Delete(
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
				if err := s.scheduler.Events().Delete(ctx, event); err != nil {
					return errors.Wrapf(
						err,
						"error deleting event %q from scheduler",
						event.ID,
					)
				}
				eventRefList.Items = append(
					eventRefList.Items,
					brignext.EventReference{
						TypeMeta: brignext.TypeMeta{
							APIVersion: brignext.APIVersion,
							Kind:       "EventReference",
						},
						ID: event.ID,
					},
				)
			}
			return nil
		}); err != nil {
			return eventRefList, err
		}
	}

	return eventRefList, nil
}

func (s *service) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	if err := s.store.Events().UpdateWorkerStatus(
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
) (brignext.LogEntryList, error) {
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
) (brignext.LogEntryList, error) {
	return s.logStore.GetWorkerInitLogs(ctx, eventID)
}

func (s *service) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamWorkerInitLogs(ctx, eventID)
}

func (s *service) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	if err := s.store.Events().UpdateJobStatus(
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
) (brignext.LogEntryList, error) {
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
) (brignext.LogEntryList, error) {
	return s.logStore.GetJobInitLogs(ctx, eventID, jobName)
}

func (s *service) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return s.logStore.StreamJobInitLogs(ctx, eventID, jobName)
}
