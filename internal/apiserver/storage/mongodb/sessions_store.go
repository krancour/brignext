package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type sessionsStore struct {
	collection *mongo.Collection
}

func NewSessionsStore(database *mongo.Database) (storage.SessionsStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("sessions")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			mongo.IndexModel{
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
	return &sessionsStore{
		collection: collection,
	}, nil
}

func (s *sessionsStore) Create(
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

func (s *sessionsStore) GetByHashedOAuth2State(
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

func (s *sessionsStore) GetByHashedToken(
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

func (s *sessionsStore) Authenticate(
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

func (s *sessionsStore) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"metadata.id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return brignext.NewErrNotFound("Session", id)
	}
	return nil
}

func (s *sessionsStore) DeleteByUser(
	ctx context.Context,
	userID string,
) (int64, error) {
	res, err := s.collection.DeleteMany(ctx, bson.M{"userID": userID})
	if err != nil {
		return 0, errors.Wrapf(err, "error deleting sessions for user %q", userID)
	}
	return res.DeletedCount, nil
}
