package sessions

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/pkg/api/auth"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TODO: DRY this up
var mongodbTimeout = 5 * time.Second

type Store interface {
	Create(context.Context, auth.Session) error
	GetByHashedOAuth2State(context.Context, string) (auth.Session, error)
	GetByHashedToken(context.Context, string) (auth.Session, error)
	Authenticate(
		ctx context.Context,
		sessionID string,
		userID string,
		expires time.Time,
	) error
	Delete(context.Context, string) error
	DeleteByUser(context.Context, string) (int64, error)

	DoTx(context.Context, func(context.Context) error) error

	CheckHealth(context.Context) error
}

type store struct {
	collection *mongo.Collection
}

func NewStore(database *mongo.Database) (Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("sessions")
	if _, err := collection.Indexes().CreateMany(
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
			// Fast lookup for completing OIDC auth
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
			// Fast lookup by bearer token
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
	return &store{
		collection: collection,
	}, nil
}

func (s *store) Create(
	ctx context.Context,
	session auth.Session,
) error {
	now := time.Now()
	session.Created = &now
	if _, err := s.collection.InsertOne(ctx, session); err != nil {
		return errors.Wrapf(err, "error inserting new session %q", session.ID)
	}
	return nil
}

func (s *store) GetByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (auth.Session, error) {
	session := auth.Session{}
	res := s.collection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": hashedOAuth2State},
	)
	if res.Err() == mongo.ErrNoDocuments {
		return session, brignext.NewErrNotFound("Session", "")
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

func (s *store) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (auth.Session, error) {
	session := auth.Session{}
	res := s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return session, brignext.NewErrNotFound("Session", "")
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

func (s *store) Authenticate(
	ctx context.Context,
	sessionID string,
	userID string,
	expires time.Time,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id": sessionID,
		},
		bson.M{
			"$set": bson.M{
				"userID":        userID,
				"authenticated": time.Now(),
				"expires":       expires,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating session %q", sessionID)
	}
	if res.MatchedCount == 0 {
		return brignext.NewErrNotFound("Session", sessionID)
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"metadata.id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return brignext.NewErrNotFound("Session", id)
	}
	return nil
}

func (s *store) DeleteByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	res, err := s.collection.DeleteMany(ctx, bson.M{"userID": userID})
	if err != nil {
		return 0, errors.Wrapf(err, "error deleting sessions for user %q", userID)
	}
	return res.DeletedCount, nil
}

func (s *store) DoTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := s.collection.Database().Client().UseSession(
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
	if err := s.collection.Database().Client().Ping(
		pingCtx,
		readpref.Primary(),
	); err != nil {
		return errors.Wrap(err, "error pinging mongodb")
	}
	return nil
}
