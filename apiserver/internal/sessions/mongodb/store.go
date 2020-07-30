package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/apimachinery/auth"
	"github.com/krancour/brignext/v2/apiserver/internal/sessions"
	brignext "github.com/krancour/brignext/v2/sdk"
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
			// Fast lookup for completing OIDC auth
			//
			// TODO: CosmosDB doesn't support this partial index. We can probably live
			// without it because lookup by hashedOAuth2State should only occur at the
			// end of an OIDC authentication flow, which should be relatively
			// uncommon. We can afford for it to be relatively slow. What we should
			// probably do is make this index optional through a configuration
			// setting.
			//
			// {
			// 	Keys: bson.M{
			// 		"hashedOAuth2State": 1,
			// 	},
			// 	Options: &options.IndexOptions{
			// 		Unique: &unique,
			// 		PartialFilterExpression: bson.M{
			// 			"hashedOAuth2State": bson.M{"exists": true},
			// 		},
			// 	},
			// },
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
		return brignext.NewErrNotFound("Session", sessionID)
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting session %q", id)
	}
	if res.DeletedCount == 0 {
		return brignext.NewErrNotFound("Session", id)
	}
	return nil
}
