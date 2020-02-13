package service

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Service interface {
	CreateUser(context.Context, brignext.User) error
	GetUsers(context.Context) ([]brignext.User, error)
	GetUser(context.Context, string) (brignext.User, bool, error)
	LockUser(context.Context, string) (bool, error)
	UnlockUser(context.Context, string) (bool, error)

	CreateRootSession(context.Context) (string, error)
	CreateUserSession(context.Context) (string, string, error)
	GetSessionByOAuth2State(
		context.Context,
		string,
	) (brignext.Session, bool, error)
	GetSessionByToken(context.Context, string) (brignext.Session, bool, error)
	AuthenticateSession(
		ctx context.Context,
		sessionID string,
		userID string,
	) (bool, error)
	DeleteSession(context.Context, string) (bool, error)
	DeleteSessionsByUser(context.Context, string) (int64, error)

	CreateServiceAccount(context.Context, brignext.ServiceAccount) (string, error)
	GetServiceAccounts(context.Context) ([]brignext.ServiceAccount, error)
	GetServiceAccount(
		context.Context,
		string,
	) (brignext.ServiceAccount, bool, error)
	GetServiceAccountByToken(
		context.Context,
		string,
	) (brignext.ServiceAccount, bool, error)
	LockServiceAccount(context.Context, string) (bool, error)
	UnlockServiceAccount(context.Context, string) (string, bool, error)

	CreateProject(context.Context, brignext.Project) error
	GetProjects(context.Context) ([]brignext.Project, error)
	GetProject(context.Context, string) (brignext.Project, bool, error)
	UpdateProject(context.Context, brignext.Project) (bool, error)
	DeleteProject(context.Context, string) (bool, error)

	CreateEvent(context.Context, brignext.Event) (string, bool, error)
	GetEvents(context.Context) ([]brignext.Event, error)
	GetEventsByProject(context.Context, string) ([]brignext.Event, error)
	GetEvent(context.Context, string) (brignext.Event, bool, error)
	DeleteEvent(
		ctx context.Context,
		id string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (bool, error)
	DeleteEventsByProject(
		ctx context.Context,
		projectID string,
		deleteAccepted bool,
		deleteProcessing bool,
	) (int64, error)
}

type service struct {
	store    storage.Store
	logStore storage.LogStore
}

func NewService(store storage.Store, logStore storage.LogStore) Service {
	return &service{
		store:    store,
		logStore: logStore,
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
) (brignext.User, bool, error) {
	user, ok, err := s.store.GetUser(ctx, id)
	if err != nil {
		return user, false, errors.Wrapf(
			err,
			"error retrieving user %q from store",
			id,
		)
	}
	return user, ok, nil
}

func (s *service) LockUser(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := s.store.DoTx(ctx, func(ctx context.Context) error {

		var err error
		if ok, err = s.store.LockUser(ctx, id); err != nil {
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
	return ok, err
}

func (s *service) UnlockUser(ctx context.Context, id string) (bool, error) {
	ok, err := s.store.UnlockUser(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "error unlocking user %q in store", id)
	}
	return ok, nil
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
) (brignext.Session, bool, error) {
	session, ok, err := s.store.GetSessionByHashedOAuth2State(
		ctx,
		crypto.ShortSHA("", oauth2State),
	)
	if err != nil {
		return session, false, errors.Wrap(
			err,
			"error retrieving session from store by hashed oauth2 state",
		)
	}
	return session, ok, nil
}

func (s *service) GetSessionByToken(
	ctx context.Context,
	token string,
) (brignext.Session, bool, error) {
	session, ok, err := s.store.GetSessionByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return session, false, errors.Wrap(
			err,
			"error retrieving session from store by hashed token",
		)
	}
	return session, ok, nil
}

func (s *service) AuthenticateSession(
	ctx context.Context,
	sessionID string,
	userID string,
) (bool, error) {
	ok, err := s.store.AuthenticateSession(
		ctx,
		sessionID,
		userID,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error storing authentication details for session %q",
			sessionID,
		)
	}
	return ok, nil
}

func (s *service) DeleteSession(ctx context.Context, id string) (bool, error) {
	ok, err := s.store.DeleteSession(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "error removing session %q from store", id)
	}
	return ok, nil
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
) (brignext.ServiceAccount, bool, error) {
	serviceAccount, ok, err := s.store.GetServiceAccount(ctx, id)
	if err != nil {
		return serviceAccount, false, errors.Wrapf(
			err,
			"error retrieving service account %q from store",
			id,
		)
	}
	return serviceAccount, ok, nil
}

func (s *service) GetServiceAccountByToken(
	ctx context.Context,
	token string,
) (brignext.ServiceAccount, bool, error) {
	serviceAccount, ok, err := s.store.GetServiceAccountByHashedToken(
		ctx,
		crypto.ShortSHA("", token),
	)
	if err != nil {
		return serviceAccount, false, errors.Wrap(
			err,
			"error retrieving service account from store by hashed token",
		)
	}
	return serviceAccount, ok, nil
}

func (s *service) LockServiceAccount(
	ctx context.Context,
	id string,
) (bool, error) {
	ok, err := s.store.LockServiceAccount(ctx, id)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error locking service account %q in the store",
			id,
		)
	}
	return ok, nil
}

func (s *service) UnlockServiceAccount(
	ctx context.Context,
	id string,
) (string, bool, error) {
	newToken := crypto.NewToken(256)
	ok, err := s.store.UnlockServiceAccount(
		ctx,
		id,
		crypto.ShortSHA("", newToken),
	)
	if err != nil {
		return "", false, errors.Wrapf(
			err,
			"error unlocking service account %q in the store",
			id,
		)
	}
	return newToken, ok, nil
}

func (s *service) CreateProject(
	ctx context.Context,
	project brignext.Project,
) error {
	now := time.Now()
	project.Created = &now
	if err := s.store.CreateProject(ctx, project); err != nil {
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
) (brignext.Project, bool, error) {
	project, ok, err := s.store.GetProject(ctx, id)
	if err != nil {
		return project, false, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, ok, nil
}

func (s *service) UpdateProject(
	ctx context.Context,
	project brignext.Project,
) (bool, error) {
	// TODO: How can we stop the created time from being overwritten?
	ok, err := s.store.UpdateProject(ctx, project)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error updating project %q in store",
			project.ID,
		)
	}
	return ok, nil
}

func (s *service) DeleteProject(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := s.store.DoTx(ctx, func(ctx context.Context) error {

		var err error
		if ok, err = s.store.DeleteProject(ctx, id); err != nil {
			return errors.Wrapf(err, "error removing project %q from store", id)
		}

		if _, err := s.store.DeleteEventsByProject(
			ctx,
			id,
			true,
			true,
		); err != nil {
			return errors.Wrapf(
				err,
				"error removing events for project %q from store",
				id,
			)
		}

		return nil
	})
	return ok, err
}

func (s *service) CreateEvent(
	ctx context.Context,
	event brignext.Event,
) (string, bool, error) {
	project, ok, err := s.store.GetProject(ctx, event.ProjectID)
	if err != nil {
		return "", false, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	} else if !ok {
		return "", false, nil
	}

	event.ID = uuid.NewV4().String()
	event.Workers = project.GetWorkers(event.Provider, event.Type)
	if len(event.Workers) == 0 {
		event.Status = brignext.EventStatusMoot
	} else {
		event.Status = brignext.EventStatusAccepted
	}
	now := time.Now()
	event.Created = &now

	if err := s.store.CreateEvent(ctx, event); err != nil {
		return "", false, errors.Wrapf(err, "error storing new event %q", event.ID)
	}

	return event.ID, true, nil
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
) (brignext.Event, bool, error) {
	event, ok, err := s.store.GetEvent(ctx, id)
	if err != nil {
		return event, false, errors.Wrapf(
			err,
			"error retrieving event %q from store",
			id,
		)
	}
	return event, ok, nil
}

func (s *service) DeleteEvent(
	ctx context.Context,
	id string,
	deleteAccepted bool,
	deleteProcessing bool,
) (bool, error) {
	ok, err := s.store.DeleteEvent(ctx, id, deleteAccepted, deleteProcessing)
	if err != nil {
		return false, errors.Wrapf(err, "error removing event %q from store", id)
	}
	return ok, nil
}

func (s *service) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
	deleteAccepted bool,
	deleteProcessing bool,
) (int64, error) {
	n, err := s.store.DeleteEventsByProject(
		ctx,
		projectID,
		deleteAccepted,
		deleteProcessing,
	)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error removing events for project %q from store",
			projectID,
		)
	}
	return n, nil
}
