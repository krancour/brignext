package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type store struct {
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

	sessionsCollection := database.Collection("sessions")
	if _, err := sessionsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"hashedOAuth2State": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
					PartialFilterExpression: bson.M{
						"hashedOAuth2State": bson.M{"exists": true},
					},
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
		return nil, errors.Wrap(err, "error adding indexes to sessions collection")
	}

	serviceAccountsCollection := database.Collection("service-accounts")
	if _, err := serviceAccountsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"hashedToken": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to service accounts collection",
		)
	}

	projectsCollection := database.Collection("projects")
	if _, err := projectsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"tags": 1,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to projects collection")
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
		database:                  database,
		usersCollection:           database.Collection("users"),
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
	findOptions.SetSort(bson.M{"_id": 1})
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
	res := s.usersCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, &brignext.ErrUserNotFound{id}
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
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"locked": true},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrUserNotFound{id}
	}
	return nil
}

func (s *store) UnlockUser(ctx context.Context, id string) error {
	res, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$unset": bson.M{"locked": 1},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrUserNotFound{id}
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
		bson.M{"hashedOAuth2State": hashedOAuth2State},
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
	res := s.sessionsCollection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
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
			"_id": sessionID,
		},
		bson.M{
			"$set": bson.M{
				"userID":        userID,
				"authenticated": true,
				"expires":       expires,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating session %q", sessionID)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrSessionNotFound{sessionID}
	}
	return nil
}

func (s *store) DeleteSession(ctx context.Context, id string) error {
	res, err := s.sessionsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return &brignext.ErrSessionNotFound{id}
	}
	return nil
}

func (s *store) DeleteSessionsByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	res, err := s.sessionsCollection.DeleteMany(ctx, bson.M{"userID": userID})
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

func (s *store) GetServiceAccounts(ctx context.Context) ([]brignext.ServiceAccount, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
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
	res := s.serviceAccountsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, &brignext.ErrServiceAccountNotFound{id}
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
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"locked": true},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrServiceAccountNotFound{id}
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
		bson.M{"_id": id},
		bson.M{
			"$unset": bson.M{"locked": 1},
			"$set":   bson.M{"hashedToken": newHashedToken},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrServiceAccountNotFound{id}
	}
	return nil
}

func (s *store) CreateProject(ctx context.Context, project brignext.Project) error {
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
	findOptions.SetSort(bson.M{"_id": 1})
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

func (s *store) GetProjectsByTags(
	ctx context.Context,
	tags brignext.ProjectTags,
) ([]brignext.Project, error) {
	conditions := make([]bson.M, len(tags))
	var i int
	for key, value := range tags {
		conditions[i] = bson.M{
			"tags": bson.M{
				"key":   key,
				"value": value,
			},
		}
		i++
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := s.projectsCollection.Find(
		ctx,
		bson.M{"$and": conditions},
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
	res := s.projectsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, &brignext.ErrProjectNotFound{id}
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
			"_id": project.ID,
		},
		bson.M{
			"$set": bson.M{
				"description":   project.Description,
				"tags":          project.Tags,
				"workerConfigs": project.WorkerConfigs,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error replacing project %q", project.ID)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrProjectNotFound{project.ID}
	}
	return nil
}

func (s *store) DeleteProject(ctx context.Context, id string) error {
	res, err := s.projectsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return &brignext.ErrProjectNotFound{id}
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
		return event, &brignext.ErrEventNotFound{id}
	}
	if res.Err() != nil {
		return event, errors.Wrapf(res.Err(), "error finding event %q", id)
	}
	if err := res.Decode(&event); err != nil {
		return event, errors.Wrapf(err, "error decoding event %q", id)
	}
	return event, nil
}

func (s *store) UpdateEventStatus(
	ctx context.Context,
	id string,
	status brignext.EventStatus,
) error {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": id,
		},
		bson.M{
			"$set": bson.M{"status": status},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating event %q status", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrEventNotFound{id}
	}
	return nil
}

func (s *store) CancelEvent(
	ctx context.Context,
	id string,
	cancelProcessing bool,
) (bool, error) {
	event, err := s.GetEvent(ctx, id)
	if err != nil {
		return false, err
	}
	phasesToCancel := []brignext.EventPhase{
		brignext.EventPhasePending,
	}
	if cancelProcessing {
		phasesToCancel = append(phasesToCancel, brignext.EventPhaseProcessing)
	}

	if event.Status.Phase == brignext.EventPhasePending {
		event.Status.Phase = brignext.EventPhaseCanceled
	} else if cancelProcessing &&
		event.Status.Phase == brignext.EventPhaseProcessing {
		event.Status.Phase = brignext.EventPhaseAborted
	} else {
		return false, nil
	}

	for workerName, worker := range event.Workers {
		if worker.Status.Phase == brignext.WorkerPhasePending {
			worker.Status.Phase = brignext.WorkerPhaseCanceled
			event.Workers[workerName] = worker
		} else if cancelProcessing &&
			worker.Status.Phase == brignext.WorkerPhaseRunning {
			worker.Status.Phase = brignext.WorkerPhaseAborted
			// There may be running jobs that need to be recorded as aborted
			for jobName, job := range worker.Jobs {
				if job.Status.Phase == brignext.JobPhaseRunning {
					job.Status.Phase = brignext.JobPhaseAborted
					worker.Jobs[jobName] = job
				}
			}
			event.Workers[workerName] = worker
		}
	}

	if _, err = s.eventsCollection.ReplaceOne(
		ctx,
		bson.M{
			"_id": id,
		},
		event,
	); err != nil {
		return false, errors.Wrapf(err, "error canceling event %q", id)
	}

	return true, nil
}

func (s *store) DeleteEvent(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteProcessing bool,
) (bool, error) {
	if _, err := s.GetEvent(ctx, id); err != nil {
		return false, err
	}
	phasesToDelete := []brignext.EventPhase{
		brignext.EventPhaseMoot,
		brignext.EventPhaseCanceled,
		brignext.EventPhaseAborted,
		brignext.EventPhaseSucceeded,
		brignext.EventPhaseFailed,
	}
	if deletePending {
		phasesToDelete = append(phasesToDelete, brignext.EventPhasePending)
	}
	if deleteProcessing {
		phasesToDelete = append(phasesToDelete, brignext.EventPhaseProcessing)
	}
	res, err := s.eventsCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":          id,
			"status.phase": bson.M{"$in": phasesToDelete},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error deleting event %q", id)
	}
	return res.DeletedCount == 1, nil
}

func (s *store) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
) error {
	if _, err := s.eventsCollection.DeleteMany(
		ctx,
		bson.M{"projectID": projectID},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting events for project %q",
			projectID,
		)
	}
	return nil
}

func (s *store) GetWorker(
	ctx context.Context,
	eventID string,
	workerName string,
) (brignext.Worker, error) {
	worker := brignext.Worker{}
	event, err := s.GetEvent(ctx, eventID)
	if err != nil {
		return worker, errors.Wrapf(err, "error finding event %q", eventID)
	}
	var ok bool
	worker, ok = event.Workers[workerName]
	if !ok {
		return worker, &brignext.ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	return worker, nil
}

func (s *store) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	status brignext.WorkerStatus,
) error {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":                                 eventID,
			fmt.Sprintf("workers.%s", workerName): bson.M{"$exists": true},
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("workers.%s.status", workerName): status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status on worker %q of event %q",
			workerName,
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	return nil
}

func (s *store) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
	status brignext.JobStatus,
) error {
	res, err := s.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":                                 eventID,
			fmt.Sprintf("workers.%s", workerName): bson.M{"$exists": true},
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("workers.%s.jobs.%s", workerName, jobName): brignext.Job{
					Status: status,
				},
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status on worker %q job %q of event %q",
			workerName,
			jobName,
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrWorkerNotFound{
			EventID:    eventID,
			WorkerName: workerName,
		}
	}
	return nil
}
