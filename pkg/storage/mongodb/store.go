package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
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
		projectsCollection:        database.Collection("projects"),
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

func (s *store) GetUser(ctx context.Context, id string) (brignext.User, bool, error) {
	user := brignext.User{}
	res := s.usersCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, false, nil
	}
	if res.Err() != nil {
		return user, false, errors.Wrapf(res.Err(), "error finding user %q", id)
	}
	if err := res.Decode(&user); err != nil {
		return user, false, errors.Wrapf(err, "error decoding user %q", id)
	}
	return user, true, nil
}

func (s *store) LockUser(ctx context.Context, id string) (bool, error) {
	res, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"locked": true},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error updating user %q", id)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) UnlockUser(ctx context.Context, id string) (bool, error) {
	res, err := s.usersCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$unset": bson.M{"locked": 1},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error updating user %q", id)
	}
	return res.MatchedCount == 1, nil
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
) (brignext.Session, bool, error) {
	session := brignext.Session{}
	res := s.sessionsCollection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": hashedOAuth2State},
	)
	if res.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if res.Err() != nil {
		return session, false, errors.Wrap(
			res.Err(),
			"error finding session by hashed OAuth2 state",
		)
	}
	if err := res.Decode(&session); err != nil {
		return session, false, errors.Wrap(err, "error decoding session")
	}
	return session, true, nil
}

func (s *store) GetSessionByHashedToken(
	ctx context.Context,
	hashedToken string,
) (brignext.Session, bool, error) {
	session := brignext.Session{}
	res := s.sessionsCollection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if res.Err() != nil {
		return session, false, errors.Wrap(
			res.Err(),
			"error finding session by hashed token",
		)
	}
	if err := res.Decode(&session); err != nil {
		return session, false, errors.Wrap(err, "error decoding session")
	}
	return session, true, nil
}

func (s *store) AuthenticateSession(
	ctx context.Context,
	sessionID string,
	userID string,
	expires time.Time,
) (bool, error) {
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
		return false, errors.Wrapf(err, "error updating session %q", sessionID)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) DeleteSession(ctx context.Context, id string) (bool, error) {
	res, err := s.sessionsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, errors.Wrapf(err, "error deleting session %q", id)
	}
	return res.DeletedCount == 1, nil
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
) (brignext.ServiceAccount, bool, error) {
	serviceAccount := brignext.ServiceAccount{}
	res := s.serviceAccountsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if res.Err() != nil {
		return serviceAccount, false, errors.Wrapf(
			res.Err(),
			"error finding service account %q",
			id,
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrapf(
			err,
			"error decoding service account %q",
			id,
		)
	}
	return serviceAccount, true, nil
}

func (s *store) GetServiceAccountByHashedToken(
	ctx context.Context,
	hashedToken string,
) (brignext.ServiceAccount, bool, error) {
	serviceAccount := brignext.ServiceAccount{}
	res :=
		s.serviceAccountsCollection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if res.Err() != nil {
		return serviceAccount, false, errors.Wrap(
			res.Err(),
			"error finding service account by hashed token",
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrap(
			err,
			"error decoding service account",
		)
	}
	return serviceAccount, true, nil
}

func (s *store) LockServiceAccount(
	ctx context.Context,
	id string,
) (bool, error) {
	res, err := s.serviceAccountsCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{"locked": true},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error updating service account %q", id)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) UnlockServiceAccount(
	ctx context.Context,
	id string,
	newHashedToken string,
) (bool, error) {
	res, err := s.serviceAccountsCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$unset": bson.M{"locked": 1},
			"$set":   bson.M{"hashedToken": newHashedToken},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error updating service account %q", id)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) CreateProject(ctx context.Context, project brignext.Project) error {
	if _, err := s.projectsCollection.InsertOne(ctx, project); err != nil {
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

func (s *store) GetProject(
	ctx context.Context,
	id string,
) (brignext.Project, bool, error) {
	project := brignext.Project{}
	res := s.projectsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, false, nil
	}
	if res.Err() != nil {
		return project, false, errors.Wrapf(
			res.Err(),
			"error finding project %q",
			id,
		)
	}
	if err := res.Decode(&project); err != nil {
		return project, false, errors.Wrapf(err, "error decoding project %q", id)
	}
	return project, true, nil
}

func (s *store) UpdateProject(
	ctx context.Context, project brignext.Project,
) (bool, error) {
	res, err := s.projectsCollection.ReplaceOne(
		ctx,
		bson.M{
			"_id": project.ID,
		},
		project,
	)
	if err != nil {
		return false, errors.Wrapf(err, "error replacing project %q", project.ID)
	}
	return res.MatchedCount == 1, nil
}

func (s *store) DeleteProject(ctx context.Context, id string) (bool, error) {
	res, err := s.projectsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, errors.Wrapf(err, "error deleting project %q", id)
	}
	return res.DeletedCount == 1, nil
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
) (brignext.Event, bool, error) {
	event := brignext.Event{}
	res := s.eventsCollection.FindOne(ctx, bson.M{"_id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return event, false, nil
	}
	if res.Err() != nil {
		return event, false, errors.Wrapf(
			res.Err(),
			"error finding event %q",
			id,
		)
	}
	if err := res.Decode(&event); err != nil {
		return event, false, errors.Wrapf(err, "error decoding event %q", id)
	}
	return event, true, nil
}

func (s *store) DeleteEvent(
	ctx context.Context,
	id string,
	deleteAccepted bool,
	deleteProcessing bool,
) (bool, error) {
	statusesToDelete := []brignext.EventStatus{
		brignext.EventStatusMoot,
		brignext.EventStatusCanceled,
		brignext.EventStatusAborted,
		brignext.EventStatusSucceeded,
		brignext.EventStatusFailed,
	}
	if deleteAccepted {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusAccepted)
	}
	if deleteProcessing {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusProcessing)
	}
	res, err := s.eventsCollection.DeleteOne(
		ctx,
		bson.M{
			"_id":    id,
			"status": bson.M{"$in": statusesToDelete},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error deleting events %q", id)
	}
	return res.DeletedCount == 1, nil
}

func (s *store) DeleteEventsByProject(
	ctx context.Context,
	projectID string,
	deleteAccepted bool,
	deleteProcessing bool,
) (int64, error) {
	statusesToDelete := []brignext.EventStatus{
		brignext.EventStatusMoot,
		brignext.EventStatusCanceled,
		brignext.EventStatusAborted,
		brignext.EventStatusSucceeded,
		brignext.EventStatusFailed,
	}
	if deleteAccepted {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusAccepted)
	}
	if deleteProcessing {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusProcessing)
	}
	res, err := s.eventsCollection.DeleteMany(
		ctx,
		bson.M{
			"projectID": projectID,
			"status":    bson.M{"$in": statusesToDelete},
		},
	)
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"error deleting events for project %q",
			projectID,
		)
	}
	return res.DeletedCount, nil
}
