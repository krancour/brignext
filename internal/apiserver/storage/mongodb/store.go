package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type store struct {
	id                        string
	database                  *mongo.Database
	usersCollection           *mongo.Collection
	sessionsCollection        *mongo.Collection
	serviceAccountsCollection *mongo.Collection
	projectsCollection        *mongo.Collection
	eventsCollection          *mongo.Collection
}

func NewStore(database *mongo.Database) (storage.Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	unique := true

	usersCollection := database.Collection("users")
	if _, err := usersCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"metadata.id": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}

	sessionsCollection := database.Collection("sessions")
	if _, err := sessionsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"metadata.id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			{
				Keys: bson.M{
					"spec.hashedOAuth2State": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
					PartialFilterExpression: bson.M{
						"spec.hashedOAuth2State": bson.M{"exists": true},
					},
				},
			},
			{
				Keys: bson.M{
					"spec.hashedToken": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to sessions collection")
	}

	serviceAccountsCollection := database.Collection("service-accounts")
	if _, err := serviceAccountsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"metadata.id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			{
				Keys: bson.M{
					"hashedToken": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to service accounts collection",
		)
	}

	projectsCollection := database.Collection("projects")
	if _, err := projectsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"metadata.id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			{
				Keys: bson.M{
					"eventSubscriptions.source": 1,
					"eventSubscriptions.types":  1,
				},
			},
			{
				Keys: bson.M{
					"eventSubscriptions.labels": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to projects collection",
		)
	}

	eventsCollection := database.Collection("events")
	if _, err := eventsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"projectID": 1,
				},
			},
			{
				Keys: bson.M{
					"created": -1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}

	return &store{
		id:                        uuid.NewV4().String(),
		database:                  database,
		usersCollection:           usersCollection,
		sessionsCollection:        sessionsCollection,
		serviceAccountsCollection: serviceAccountsCollection,
		projectsCollection:        projectsCollection,
		eventsCollection:          eventsCollection,
	}, nil
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

func (s *store) CreateUser(ctx context.Context, user brignext.User) error {
	if _, err :=
		s.usersCollection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &brignext.ErrUserIDConflict{
					ID: user.ID,
				}
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (s *store) GetUsers(ctx context.Context) ([]brignext.User, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.usersCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error finding users")
	}
	users := []brignext.User{}
	if err := cur.All(ctx, &users); err != nil {
		return nil, errors.Wrap(err, "error decoding users")
	}
	return users, nil
}

func (s *store) GetUser(ctx context.Context, id string) (brignext.User, error) {
	user := brignext.User{}
	res := s.usersCollection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, &brignext.ErrUserNotFound{
			ID: id,
		}
	}
	if res.Err() != nil {
		return user, errors.Wrapf(res.Err(), "error finding user %q", id)
	}
	if err := res.Decode(&user); err != nil {
		return user, errors.Wrapf(err, "error decoding user %q", id)
	}
	return user, nil
}

func (s *store) LockUser(ctx context.Context, id string) error {
	res, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
		bson.M{
			"$set": bson.M{"status.locked": true},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrUserNotFound{
			ID: id,
		}
	}
	return nil
}

func (s *store) UnlockUser(ctx context.Context, id string) error {
	res, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
		bson.M{
			"$set": bson.M{"status.locked": false},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrUserNotFound{
			ID: id,
		}
	}
	return nil
}

func (s *store) CreateSession(
	ctx context.Context,
	session brignext.Session,
) error {
	if _, err := s.sessionsCollection.InsertOne(ctx, session); err != nil {
		return errors.Wrapf(err, "error inserting new session %q", session.ID)
	}
	return nil
}

func (s *store) GetSessionByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (brignext.Session, error) {
	session := brignext.Session{}
	res := s.sessionsCollection.FindOne(
		ctx,
		bson.M{"spec.hashedOAuth2State": hashedOAuth2State},
	)
	if res.Err() == mongo.ErrNoDocuments {
		return session, &brignext.ErrSessionNotFound{}
	}
	if res.Err() != nil {
		return session, errors.Wrap(
			res.Err(),
			"error finding session by hashed OAuth2 state",
		)
	}
	if err := res.Decode(&session); err != nil {
		return session, errors.Wrap(err, "error decoding session")
	}
	return session, nil
}

func (s *store) GetSessionByHashedToken(
	ctx context.Context,
	hashedToken string,
) (brignext.Session, error) {
	session := brignext.Session{}
	res := s.sessionsCollection.FindOne(ctx, bson.M{"spec.hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return session, &brignext.ErrSessionNotFound{}
	}
	if res.Err() != nil {
		return session, errors.Wrap(
			res.Err(),
			"error finding session by hashed token",
		)
	}
	if err := res.Decode(&session); err != nil {
		return session, errors.Wrap(err, "error decoding session")
	}
	return session, nil
}

func (s *store) AuthenticateSession(
	ctx context.Context,
	sessionID string,
	userID string,
	expires time.Time,
) error {
	res, err := s.sessionsCollection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id": sessionID,
		},
		bson.M{
			"$set": bson.M{
				"metadata.expires":     expires,
				"spec.userID":          userID,
				"status.authenticated": true,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating session %q", sessionID)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrSessionNotFound{
			ID: sessionID,
		}
	}
	return nil
}

func (s *store) DeleteSession(ctx context.Context, id string) error {
	res, err := s.sessionsCollection.DeleteOne(ctx, bson.M{"metadata.id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return &brignext.ErrSessionNotFound{
			ID: id,
		}
	}
	return nil
}

func (s *store) DeleteSessionsByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	res, err := s.sessionsCollection.DeleteMany(ctx, bson.M{"spec.userID": userID})
	if err != nil {
		return 0, errors.Wrapf(err, "error deleting sessions for user %q", userID)
	}
	return res.DeletedCount, nil
}

func (s *store) CreateServiceAccount(
	ctx context.Context,
	serviceAccount brignext.ServiceAccount,
) error {
	if _, err := s.serviceAccountsCollection.InsertOne(
		ctx,
		serviceAccount,
	); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &brignext.ErrServiceAccountIDConflict{
					ID: serviceAccount.ID,
				}
			}
		}
		return errors.Wrapf(
			err,
			"error inserting new service account %q",
			serviceAccount.ID,
		)
	}
	return nil
}

func (s *store) GetServiceAccounts(
	ctx context.Context,
) ([]brignext.ServiceAccount, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.serviceAccountsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error finding service accounts")
	}
	serviceAccounts := []brignext.ServiceAccount{}
	if err := cur.All(ctx, &serviceAccounts); err != nil {
		return nil, errors.Wrap(err, "error decoding service accounts")
	}
	return serviceAccounts, nil
}

func (s *store) GetServiceAccount(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	res := s.serviceAccountsCollection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, &brignext.ErrServiceAccountNotFound{
			ID: id,
		}
	}
	if res.Err() != nil {
		return serviceAccount, errors.Wrapf(
			res.Err(),
			"error finding service account %q",
			id,
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error decoding service account %q",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *store) GetServiceAccountByHashedToken(
	ctx context.Context,
	hashedToken string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	res :=
		s.serviceAccountsCollection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, &brignext.ErrServiceAccountNotFound{}
	}
	if res.Err() != nil {
		return serviceAccount, errors.Wrap(
			res.Err(),
			"error finding service account by hashed token",
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, errors.Wrap(
			err,
			"error decoding service account",
		)
	}
	return serviceAccount, nil
}

func (s *store) LockServiceAccount(
	ctx context.Context,
	id string,
) error {
	res, err := s.serviceAccountsCollection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
		bson.M{
			"$set": bson.M{"status.locked": true},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrServiceAccountNotFound{
			ID: id,
		}
	}
	return nil
}

func (s *store) UnlockServiceAccount(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	res, err := s.serviceAccountsCollection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
		bson.M{
			"$set": bson.M{
				"status.locked": false,
				"hashedToken":   newHashedToken,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrServiceAccountNotFound{
			ID: id,
		}
	}
	return nil
}

func (s *store) CreateProject(
	ctx context.Context,
	project brignext.Project,
) error {
	if _, err := s.projectsCollection.InsertOne(ctx, project); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &brignext.ErrProjectIDConflict{
					ID: project.ID,
				}
			}
		}
		return errors.Wrapf(err, "error inserting new project %q", project.ID)
	}
	return nil
}

func (s *store) GetProjects(ctx context.Context) ([]brignext.Project, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.projectsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error finding projects")
	}
	projects := []brignext.Project{}
	if err := cur.All(ctx, &projects); err != nil {
		return nil, errors.Wrap(err, "error decoding projects")
	}
	return projects, nil
}

func (s *store) GetSubscribedProjects(
	ctx context.Context,
	event brignext.Event,
) ([]brignext.Project, error) {
	subscriptionMatchCriteria := bson.M{
		"source": event.Source,
		"types": bson.M{
			"$in": []string{event.Type, "*"},
		},
	}
	if len(event.Labels) > 0 {
		labelConditions := make([]bson.M, len(event.Labels))
		var i int
		for key, value := range event.Labels {
			labelConditions[i] = bson.M{
				"$elemMatch": bson.M{
					"key":   key,
					"value": value,
				},
			}
			i++
		}
		subscriptionMatchCriteria["labels"] = bson.M{
			"$all": labelConditions,
		}
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := s.projectsCollection.Find(
		ctx,
		bson.M{
			"spec.eventSubscriptions": bson.M{
				"$elemMatch": subscriptionMatchCriteria,
			},
		},
		findOptions,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error finding projects")
	}
	projects := []brignext.Project{}
	if err := cur.All(ctx, &projects); err != nil {
		return nil, errors.Wrap(err, "error decoding projects")
	}
	return projects, nil
}

func (s *store) GetProject(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project := brignext.Project{}
	res := s.projectsCollection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, &brignext.ErrProjectNotFound{
			ID: id,
		}
	}
	if res.Err() != nil {
		return project, errors.Wrapf(res.Err(), "error finding project %q", id)
	}
	if err := res.Decode(&project); err != nil {
		return project, errors.Wrapf(err, "error decoding project %q", id)
	}
	return project, nil
}

func (s *store) UpdateProject(
	ctx context.Context, project brignext.Project,
) error {
	res, err := s.projectsCollection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id": project.ID,
		},
		bson.M{
			"$set": bson.M{
				"spec": project.Spec,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error replacing project %q", project.ID)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrProjectNotFound{
			ID: project.ID,
		}
	}
	return nil
}

func (s *store) DeleteProject(ctx context.Context, id string) error {
	res, err := s.projectsCollection.DeleteOne(ctx, bson.M{"metadata.id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return &brignext.ErrProjectNotFound{
			ID: id,
		}
	}
	if _, err := s.eventsCollection.DeleteMany(
		ctx,
		bson.M{"projectID": id},
	); err != nil {
		return errors.Wrapf(err, "error deleting events for project %q", id)
	}
	return nil
}

func (s *store) CreateEvent(ctx context.Context, event brignext.Event) error {
	if _, err := s.eventsCollection.InsertOne(ctx, event); err != nil {
		return errors.Wrapf(err, "error inserting new event %q", event.ID)
	}
	return nil
}

func (s *store) GetEvents(ctx context.Context) ([]brignext.Event, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err := s.eventsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error finding events")
	}
	events := []brignext.Event{}
	if err := cur.All(ctx, &events); err != nil {
		return nil, errors.Wrap(err, "error decoding events")
	}
	return events, nil
}

func (s *store) GetEventsByProject(
	ctx context.Context,
	projectID string,
) ([]brignext.Event, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err :=
		s.eventsCollection.Find(ctx, bson.M{"projectID": projectID}, findOptions)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error finding events for project %q",
			projectID,
		)
	}
	events := []brignext.Event{}
	if err := cur.All(ctx, &events); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding events for project %q",
			projectID,
		)
	}
	return events, nil
}

func (s *store) GetEvent(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event := brignext.Event{}
	res := s.eventsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return event, &brignext.ErrEventNotFound{
			ID: id,
		}
	}
	if res.Err() != nil {
		return event, errors.Wrapf(res.Err(), "error finding event %q", id)
	}
	if err := res.Decode(&event); err != nil {
		return event, errors.Wrapf(err, "error decoding event %q", id)
	}
	return event, nil
}

func (s *store) CancelEvent(
	ctx context.Context,
	id string,
	cancelRunning bool,
) (bool, error) {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":                 id,
			"worker.status.phase": brignext.WorkerPhasePending,
		},
		bson.M{
			"$set": bson.M{
				"worker.status.phase": brignext.WorkerPhaseCanceled,
			},
		},
	)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error updating status of event %q worker",
			id,
		)
	}
	if res.MatchedCount == 1 {
		return true, nil
	}

	if !cancelRunning {
		return false, nil
	}

	res, err = s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":                 id,
			"worker.status.phase": brignext.WorkerPhaseRunning,
		},
		bson.M{
			"$set": bson.M{
				"worker.status.phase": brignext.WorkerPhaseAborted,
			},
		},
	)
	if err != nil {
		return false, errors.Wrapf(
			err,
			"error updating status of event %q worker",
			id,
		)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
) (bool, error) {
	if _, err := s.GetEvent(ctx, id); err != nil {
		return false, err
	}
	phasesToDelete := []brignext.WorkerPhase{
		brignext.WorkerPhaseCanceled,
		brignext.WorkerPhaseAborted,
		brignext.WorkerPhaseSucceeded,
		brignext.WorkerPhaseFailed,
		brignext.WorkerPhaseTimedOut,
	}
	if deletePending {
		phasesToDelete = append(phasesToDelete, brignext.WorkerPhasePending)
	}
	if deleteRunning {
		phasesToDelete = append(phasesToDelete, brignext.WorkerPhaseRunning)
	}
	res, err := s.eventsCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":                 id,
			"worker.status.phase": bson.M{"$in": phasesToDelete},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error deleting event %q", id)
	}
	return res.DeletedCount == 1, nil
}

func (s *store) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{"_id": eventID},
		bson.M{
			"$set": bson.M{
				"worker.status": status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker",
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrEventNotFound{
			ID: eventID,
		}
	}
	return nil
}

func (s *store) GetJob(
	ctx context.Context,
	eventID string,
	jobName string,
) (brignext.Job, error) {
	job := brignext.Job{}
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return job, err
	}
	var ok bool
	job, ok = event.Worker.Jobs[jobName]
	if !ok {
		return job, &brignext.ErrJobNotFound{
			EventID: eventID,
			JobName: jobName,
		}
	}
	return job, nil
}

func (s *store) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": eventID,
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s", jobName): brignext.Job{
					Status: status,
				},
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q",
			eventID,
			jobName,
		)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrEventNotFound{
			ID: eventID,
		}
	}
	return nil
}
