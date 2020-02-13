package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/logic"
	"github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
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
	workersCollection         *mongo.Collection
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

	workersCollection := database.Collection("workers")
	if _, err := workersCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"eventID": 1,
				},
			},
			{
				Keys: bson.M{
					"projectID": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to workers collection")
	}

	return &store{
		database:                  database,
		usersCollection:           database.Collection("users"),
		sessionsCollection:        sessionsCollection,
		serviceAccountsCollection: serviceAccountsCollection,
		projectsCollection:        database.Collection("projects"),
		eventsCollection:          eventsCollection,
		workersCollection:         workersCollection,
	}, nil
}

func (s *store) CreateUser(user brignext.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	user.FirstSeen = time.Now()

	if _, err :=
		s.usersCollection.InsertOne(ctx, user); err != nil {
		return errors.Wrapf(err, "error creating user %q", user.ID)
	}

	return nil
}

func (s *store) GetUsers() ([]brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := s.usersCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users")
	}

	users := []brignext.User{}
	if err := cur.All(ctx, &users); err != nil {
		return nil, errors.Wrap(err, "error decoding users")
	}

	return users, nil
}

func (s *store) GetUser(id string) (brignext.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	user := brignext.User{}

	result := s.usersCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return user, false, nil
	}
	if result.Err() != nil {
		return user, false, errors.Wrapf(
			result.Err(),
			"error retrieving user %q",
			id,
		)
	}

	if err := result.Decode(&user); err != nil {
		return user, false, errors.Wrapf(err, "error decoding user %q", id)
	}

	return user, true, nil
}

func (s *store) LockUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	return mongodb.DoTx(ctx, s.database,
		func(sc mongo.SessionContext) error {

			if _, err :=
				s.usersCollection.UpdateOne(
					sc,
					bson.M{"_id": id},
					bson.M{
						"$set": bson.M{"locked": true},
					},
				); err != nil {
				return errors.Wrapf(err, "error locking user %q", id)
			}

			if _, err := s.sessionsCollection.DeleteMany(
				sc,
				bson.M{"userID": id},
			); err != nil {
				return errors.Wrapf(err, "error deleting user %q sessions", id)
			}

			return nil
		},
	)
}

func (s *store) UnlockUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.usersCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$unset": bson.M{"locked": 1},
			},
		); err != nil {
		return errors.Wrapf(err, "error unlocking user %q", id)
	}

	return nil
}

func (s *store) CreateSession(session brignext.Session) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session.ID = uuid.NewV4().String()
	session.Created = time.Now()

	var oauth2State, hashedOAuth2State string
	if !session.Root {
		oauth2State = crypto.NewToken(30)
		hashedOAuth2State = crypto.ShortSHA("", oauth2State)
	}

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err := s.sessionsCollection.InsertOne(
		ctx,
		struct {
			brignext.Session  `bson:",inline"`
			HashedOAuth2State string `bson:"hashedOAuth2State,omitempty"`
			HashedToken       string `bson:"hashedToken"`
		}{
			Session:           session,
			HashedOAuth2State: hashedOAuth2State,
			HashedToken:       hashedToken,
		},
	); err != nil {
		return "", "", "", errors.Wrap(err, "error creating new session")
	}

	return session.ID, oauth2State, token, nil
}

func (s *store) GetSession(
	criteria storage.GetSessionCriteria,
) (brignext.Session, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session := brignext.Session{}

	bsonCriteria := bson.M{}
	if !logic.ExactlyOne(
		criteria.OAuth2State != "",
		criteria.Token != "",
	) {
		return session, false, errors.New(
			"invalid criteria: only ONE oauth2 state OR token must be specified",
		)
	}
	if criteria.OAuth2State != "" {
		bsonCriteria["hashedOAuth2State"] =
			crypto.ShortSHA("", criteria.OAuth2State)
	} else if criteria.Token != "" {
		bsonCriteria["hashedToken"] = crypto.ShortSHA("", criteria.Token)
	}

	result := s.sessionsCollection.FindOne(ctx, bsonCriteria)
	if result.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if result.Err() != nil {
		return session, false, errors.Wrap(result.Err(), "error retrieving session")
	}
	if err := result.Decode(&session); err != nil {
		return session, false, errors.Wrap(err, "error decoding session")
	}

	return session, true, nil
}

func (s *store) AuthenticateSession(sessionID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.sessionsCollection.UpdateOne(
			ctx,
			bson.M{
				"_id": sessionID,
			},
			bson.M{
				"$set": bson.M{
					"userID":        userID,
					"authenticated": true,
					"expires":       time.Now().Add(time.Hour),
				},
			},
		); err != nil {
		return errors.Wrap(err, "error updating session")
	}

	return nil
}

func (s *store) DeleteSessions(
	criteria storage.DeleteSessionsCriteria,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	bsonCriteria := bson.M{}
	if !logic.ExactlyOne(
		criteria.SessionID != "",
		criteria.UserID != "",
	) {
		return errors.New(
			"invalid criteria: only ONE of session ID OR user ID must be specified",
		)
	}
	if criteria.SessionID != "" {
		bsonCriteria["_id"] = criteria.SessionID
	} else if criteria.UserID != "" {
		bsonCriteria["userID"] = criteria.UserID
	}

	if _, err := s.sessionsCollection.DeleteMany(ctx, bsonCriteria); err != nil {
		return errors.Wrap(err, "error deleting sessions")
	}

	return nil
}

func (s *store) CreateServiceAccount(
	serviceAccount brignext.ServiceAccount,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	now := time.Now()
	serviceAccount.Created = &now

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err :=
		s.serviceAccountsCollection.InsertOne(
			ctx,
			struct {
				brignext.ServiceAccount `bson:",inline"`
				HashedToken             string `bson:"hashedToken"`
			}{
				ServiceAccount: serviceAccount,
				HashedToken:    hashedToken,
			},
		); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating service account %q",
			serviceAccount.ID,
		)
	}

	return token, nil
}

func (s *store) GetServiceAccounts() ([]brignext.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := s.serviceAccountsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving service accounts")
	}

	serviceAccounts := []brignext.ServiceAccount{}
	if err := cur.All(ctx, &serviceAccounts); err != nil {
		return nil, errors.Wrap(err, "error decoding service accounts")
	}

	return serviceAccounts, nil
}

func (s *store) GetServiceAccount(
	criteria storage.GetServiceAccountCriteria,
) (brignext.ServiceAccount, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	serviceAccount := brignext.ServiceAccount{}

	bsonCriteria := bson.M{}
	if !logic.ExactlyOne(
		criteria.ServiceAccountID != "",
		criteria.Token != "",
	) {
		return serviceAccount, false, errors.New(
			"invalid criteria: only ONE of service account ID OR token must be " +
				"specified",
		)
	}
	if criteria.ServiceAccountID != "" {
		bsonCriteria["_id"] = criteria.ServiceAccountID
	} else if criteria.Token != "" {
		bsonCriteria["hashedToken"] = crypto.ShortSHA("", criteria.Token)
	}

	result := s.serviceAccountsCollection.FindOne(ctx, bsonCriteria)
	if result.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if result.Err() != nil {
		return serviceAccount, false, errors.Wrap(
			result.Err(),
			"error retrieving service account",
		)
	}
	if err := result.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrap(
			err,
			"error decoding service account",
		)
	}

	return serviceAccount, true, nil
}

func (s *store) LockServiceAccount(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$set": bson.M{"locked": true},
			},
		); err != nil {
		return errors.Wrapf(err, "error locking service account %q", id)
	}

	return nil
}

func (s *store) UnlockServiceAccount(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err :=
		s.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$unset": bson.M{"locked": 1},
				"$set":   bson.M{"hashedToken": hashedToken},
			},
		); err != nil {
		return "", errors.Wrapf(err, "error unlocking service account %q", id)
	}

	return token, nil
}

func (s *store) CreateProject(project brignext.Project) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	now := time.Now()
	project.Created = &now

	if _, err := s.projectsCollection.InsertOne(ctx, project); err != nil {
		return "", errors.Wrapf(err, "error creating project %q", project.ID)
	}

	return project.ID, nil
}

func (s *store) GetProjects() ([]brignext.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := s.projectsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving projects")
	}

	projects := []brignext.Project{}
	if err := cur.All(ctx, &projects); err != nil {
		return nil, errors.Wrap(err, "error decoding projects")
	}

	return projects, nil
}

func (s *store) GetProject(id string) (brignext.Project, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	project := brignext.Project{}

	result := s.projectsCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return project, false, nil
	}
	if result.Err() != nil {
		return project, false, errors.Wrapf(
			result.Err(),
			"error retrieving project %q",
			id,
		)
	}

	if err := result.Decode(&project); err != nil {
		return project, false, errors.Wrapf(err, "error decoding project %q", id)
	}

	return project, true, nil
}

func (s *store) UpdateProject(project brignext.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.projectsCollection.ReplaceOne(
			ctx,
			bson.M{
				"_id": project.ID,
			},
			project,
		); err != nil {
		return errors.Wrapf(err, "error updating project %q", project.ID)
	}

	return nil
}

func (s *store) DeleteProject(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	return mongodb.DoTx(ctx, s.database,
		func(sc mongo.SessionContext) error {

			if _, err :=
				s.projectsCollection.DeleteOne(sc, bson.M{"_id": id}); err != nil {
				return errors.Wrapf(err, "error deleting project %q", id)
			}

			if _, err :=
				s.eventsCollection.DeleteMany(sc, bson.M{"projectID": id}); err != nil {
				return errors.Wrapf(err, "error deleting events for project %q", id)
			}

			if _, err :=
				s.workersCollection.DeleteMany(
					sc,
					bson.M{"projectID": id},
				); err != nil {
				return errors.Wrapf(err, "error deleting workers for project %q", id)
			}

			return nil
		},
	)
}

func (s *store) CreateEvent(event brignext.Event) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event.ID = uuid.NewV4().String()
	workers := make([]interface{}, len(event.Workers))
	if len(event.Workers) == 0 {
		event.Status = brignext.EventStatusMoot
	} else {
		event.Status = brignext.EventStatusAccepted
		for i, worker := range event.Workers {
			worker.ID = uuid.NewV4().String()
			worker.EventID = event.ID
			worker.ProjectID = event.ProjectID
			worker.Status = brignext.WorkerStatusPending
			workers[i] = worker
		}
	}
	now := time.Now()
	event.Created = &now
	event.Workers = nil

	return event.ID, mongodb.DoTx(ctx, s.database,
		func(sc mongo.SessionContext) error {

			if _, err := s.eventsCollection.InsertOne(sc, event); err != nil {
				return errors.Wrapf(err, "error creating event %q", event.ID)
			}

			if len(workers) > 0 {
				if _, err := s.workersCollection.InsertMany(sc, workers); err != nil {
					return errors.Wrapf(
						err,
						"error creating workers for event %q",
						event.ID,
					)
				}
			}

			return nil
		},
	)
}

func (s *store) GetEvents(
	criteria storage.GetEventsCriteria,
) ([]brignext.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	bsonCriteria := bson.M{}
	if criteria.ProjectID != "" {
		bsonCriteria["projectID"] = criteria.ProjectID
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err := s.eventsCollection.Find(ctx, bsonCriteria, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving events")
	}

	events := []brignext.Event{}
	if err := cur.All(ctx, &events); err != nil {
		return nil, errors.Wrap(err, "error decoding events")
	}

	return events, nil
}

func (s *store) GetEvent(id string) (brignext.Event, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event := brignext.Event{}

	events := []brignext.Event{}
	cur, err := s.eventsCollection.Aggregate(
		ctx,
		[]bson.M{
			{
				"$match": bson.M{"_id": id},
			},
			{
				"$lookup": bson.M{ // Left outer join
					"from":         "workers",
					"localField":   "_id",
					"foreignField": "eventID",
					"as":           "workers",
				},
			},
		},
	)
	if err != nil {
		return event, false, errors.Wrapf(
			err,
			"error retrieving event %q",
			id,
		)
	}
	if err := cur.All(ctx, &events); err != nil {
		return event, false, errors.Wrapf(err, "error decoding event %q", id)
	}

	if len(events) == 0 {
		return event, false, nil
	}

	return events[0], true, nil
}

func (s *store) DeleteEvents(
	criteria storage.DeleteEventsCriteria,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if !logic.ExactlyOne(
		criteria.EventID != "",
		criteria.ProjectID != "",
	) {
		return errors.New(
			"invalid criteria: only ONE of event ID or project ID must be specified",
		)
	}

	bsonCriteria := bson.M{}
	if criteria.EventID != "" {
		bsonCriteria["_id"] = criteria.EventID
	} else if criteria.ProjectID != "" {
		bsonCriteria["projectID"] = criteria.ProjectID
	}
	statusesToDelete := []brignext.EventStatus{
		brignext.EventStatusMoot,
		brignext.EventStatusCanceled,
		brignext.EventStatusAborted,
		brignext.EventStatusSucceeded,
		brignext.EventStatusFailed,
	}
	if criteria.DeleteAcceptedEvents {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusAccepted)
	}
	if criteria.DeleteProcessingEvents {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusProcessing)
	}
	bsonCriteria["status"] = bson.M{"$in": statusesToDelete}

	if _, err := s.eventsCollection.DeleteMany(ctx, bsonCriteria); err != nil {
		return errors.Wrap(err, "error deleting events")
	}

	// TODO: Cascade the delete to orphaned workers. Not yet sure how.

	return nil
}
