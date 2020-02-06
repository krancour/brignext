package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type sessionStore struct {
	sessionsCollection *mongo.Collection
}

func NewSessionStore(database *mongo.Database) storage.SessionStore {
	return &sessionStore{
		sessionsCollection: database.Collection("sessions"),
	}
}

func (s *sessionStore) CreateSession() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	oauth2State := crypto.NewToken(30)
	token := crypto.NewToken(256)

	if _, err := s.sessionsCollection.InsertOne(
		ctx,
		bson.M{
			"id":                uuid.NewV4().String(),
			"hashedoauth2state": crypto.ShortSHA("", oauth2State),
			"hashedtoken":       crypto.ShortSHA("", token),
			"authenticated":     false,
		},
	); err != nil {
		return "", "", errors.Wrap(err, "error creating new session")
	}

	return oauth2State, token, nil
}

func (s *sessionStore) CreateRootSession() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	token := crypto.NewToken(256)

	if _, err := s.sessionsCollection.InsertOne(
		ctx,
		bson.M{
			"id":            uuid.NewV4().String(),
			"root":          true,
			"hashedtoken":   crypto.ShortSHA("", token),
			"authenticated": true,
			"expiresat":     time.Now().Add(10 * time.Minute),
		},
	); err != nil {
		return "", errors.Wrap(err, "error creating new root session")
	}

	return token, nil
}

func (s *sessionStore) GetSessionByOAuth2State(
	oauth2State string,
) (*brignext.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := s.sessionsCollection.FindOne(
		ctx,
		bson.M{
			"hashedoauth2state": crypto.ShortSHA("", oauth2State),
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrap(
			result.Err(),
			"error retrieving session by OAuth2 state",
		)
	}
	session := &brignext.Session{}
	if err := result.Decode(session); err != nil {
		return nil, errors.Wrap(
			err,
			"error decoding session",
		)
	}
	return session, nil
}

func (s *sessionStore) GetSessionByToken(
	token string,
) (*brignext.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := s.sessionsCollection.FindOne(
		ctx,
		bson.M{
			"hashedtoken": crypto.ShortSHA("", token),
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrap(
			result.Err(),
			"error retrieving session by token",
		)
	}
	session := &brignext.Session{}
	if err := result.Decode(session); err != nil {
		return nil, errors.Wrap(
			err,
			"error decoding session",
		)
	}
	return session, nil
}

func (s *sessionStore) AuthenticateSession(sessionID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		s.sessionsCollection.UpdateOne(
			ctx,
			bson.M{
				"id": sessionID,
			},
			bson.M{
				"$set": bson.M{
					"userid":        userID,
					"authenticated": true,
					"expiresat":     time.Now().Add(time.Hour),
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
		s.sessionsCollection.DeleteOne(
			ctx,
			bson.M{
				"id": id,
			},
		); err != nil {
		return errors.Wrap(
			err,
			"error deleting session",
		)
	}
	return nil
}
