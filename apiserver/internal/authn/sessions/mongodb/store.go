package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/authn/sessions"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const createIndexTimeout = 5 * time.Second

type store struct {
	collection *mongo.Collection
}

func NewStore(database *mongo.Database) (sessions.Store, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("sessions")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			// Fast lookup by token
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
	session authn.Session,
) error {
	if _, err := s.collection.InsertOne(ctx, session); err != nil {
		return errors.Wrapf(err, "error inserting new session %q", session.ID)
	}
	return nil
}

func (s *store) GetByHashedOAuth2State(
	ctx context.Context,
	hashedOAuth2State string,
) (authn.Session, error) {
	session := authn.Session{}
	res := s.collection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": hashedOAuth2State},
	)
	if res.Err() == mongo.ErrNoDocuments {
		return session, &core.ErrNotFound{
			Type: "Session",
		}
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
) (authn.Session, error) {
	session := authn.Session{}
	res := s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return session, &core.ErrNotFound{
			Type: "Session",
		}
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
			"id": sessionID,
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
		return &core.ErrNotFound{
			Type: "Session",
			ID:   sessionID,
		}
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return &core.ErrNotFound{
			Type: "Session",
			ID:   id,
		}
	}
	return nil
}
