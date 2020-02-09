package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type sessionStore struct {
	sessionsCollection *mongo.Collection
}

func NewSessionStore(database *mongo.Database) (storage.SessionStore, error) {
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

	return &sessionStore{
		sessionsCollection: sessionsCollection,
	}, nil
}

func (s *sessionStore) CreateSession(session brignext.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	var hashedOAuth2State string
	if session.OAuth2State != "" {
		hashedOAuth2State = crypto.ShortSHA("", session.OAuth2State)
	}
	// The bson struct tags should stop this clear text field from being
	// persisted, but this is here for good measure.
	session.OAuth2State = ""

	hashedToken := crypto.ShortSHA("", session.Token)
	// The bson struct tags should stop this clear text field from being
	// persisted, but this is here for good measure.
	session.Token = ""

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
		return errors.Wrap(err, "error creating new session")
	}

	return nil
}

func (s *sessionStore) GetSessionByOAuth2State(
	oauth2State string,
) (brignext.Session, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session := brignext.Session{}

	result := s.sessionsCollection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": crypto.ShortSHA("", oauth2State)},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if result.Err() != nil {
		return session, false, errors.Wrap(
			result.Err(),
			"error retrieving session with OAuth2 state [REDACTED]",
		)
	}
	if err := result.Decode(&session); err != nil {
		return session, false, errors.Wrap(
			err,
			"error decoding session with OAuth2 state [REDACTED]",
		)
	}
	return session, true, nil
}

func (s *sessionStore) GetSessionByToken(
	token string,
) (brignext.Session, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session := brignext.Session{}

	result := s.sessionsCollection.FindOne(
		ctx,
		bson.M{"hashedToken": crypto.ShortSHA("", token)},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if result.Err() != nil {
		return session, false, errors.Wrap(
			result.Err(),
			"error retrieving session with token [REDACTED]",
		)
	}
	if err := result.Decode(&session); err != nil {
		return session, false, errors.Wrap(
			err,
			"error decoding session with token [REDACTED]",
		)
	}
	return session, true, nil
}

func (s *sessionStore) AuthenticateSession(sessionID, userID string) error {
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

func (s *sessionStore) DeleteSession(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.sessionsCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return errors.Wrap(
			err,
			"error deleting session",
		)
	}
	return nil
}
