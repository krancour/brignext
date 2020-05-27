package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type store struct {
	database             *mongo.Database
	eventsStore          storage.EventsStore
	projectsStore        storage.ProjectsStore
	serviceAccountsStore storage.ServiceAccountsStore
	sessionsStore        storage.SessionsStore
	usersStore           storage.UsersStore
}

func NewStore(database *mongo.Database) (storage.Store, error) {
	sessionsStore, err := NewSessionsStore(database)
	if err != nil {
		return nil, err
	}
	usersStore, err := NewUsersStore(database)
	if err != nil {
		return nil, err
	}
	serviceAccountsStore, err := NewServiceAccountsStore(database)
	if err != nil {
		return nil, err
	}
	projectsStore, err := NewProjectsStore(database)
	if err != nil {
		return nil, err
	}
	eventsStore, err := NewEventsStore(database)
	if err != nil {
		return nil, err
	}

	return &store{
		database:             database,
		eventsStore:          eventsStore,
		projectsStore:        projectsStore,
		serviceAccountsStore: serviceAccountsStore,
		sessionsStore:        sessionsStore,
		usersStore:           usersStore,
	}, nil
}

func (s *store) Events() storage.EventsStore {
	return s.eventsStore
}

func (s *store) Projects() storage.ProjectsStore {
	return s.projectsStore
}

func (s *store) ServiceAccounts() storage.ServiceAccountsStore {
	return s.serviceAccountsStore
}

func (s *store) Sessions() storage.SessionsStore {
	return s.sessionsStore
}

func (s *store) Users() storage.UsersStore {
	return s.usersStore
}

func (s *store) DoTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := s.database.Client().UseSession(
		ctx,
		func(sc mongo.SessionContext) error {
			if err := sc.StartTransaction(); err != nil {
				return errors.Wrapf(err, "error starting transaction")
			}
			if err := fn(sc); err != nil {
				return err
			}
			if err := sc.CommitTransaction(sc); err != nil {
				return errors.Wrap(err, "error committing transaction")
			}
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}

func (s *store) CheckHealth(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.database.Client().Ping(pingCtx, readpref.Primary()); err != nil {
		return errors.Wrap(err, "error pinging mongo")
	}
	return nil
}
