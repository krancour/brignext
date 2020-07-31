package mongodb

import (
	"context"
	"fmt"
	"time"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/users"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const createIndexTimeout = 5 * time.Second

type store struct {
	collection         *mongo.Collection
	sessionsCollection *mongo.Collection
}

func NewStore(database *mongo.Database) (users.Store, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("users")
	if _, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"id": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}
	return &store{
		collection:         collection,
		sessionsCollection: database.Collection("sessions"),
	}, nil
}

func (s *store) Create(ctx context.Context, user brignext.User) error {
	if _, err :=
		s.collection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &brignext.ErrConflict{
					Type:   "User",
					ID:     user.ID,
					Reason: fmt.Sprintf("A user with the ID %q already exists.", user.ID),
				}
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (s *store) List(ctx context.Context) (brignext.UserReferenceList, error) {
	userList := brignext.UserReferenceList{}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	cur, err := s.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return userList, errors.Wrap(err, "error finding users")
	}
	if err := cur.All(ctx, &userList.Items); err != nil {
		return userList, errors.Wrap(err, "error decoding users")
	}
	return userList, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user := brignext.User{}
	res := s.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, &brignext.ErrNotFound{
			Type: "User",
			ID:   id,
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

func (s *store) Lock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": time.Now(),
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}

	// Now delete all the user's sessions. Note we're deliberately not doing this
	// in a transaction. This way if an error occurs after successfully locking
	// the user, but BEFORE OR WHILE deleting their existing sessions, at least
	// the user will be locked.
	if _, err = s.sessionsCollection.DeleteMany(
		ctx,
		bson.M{
			"userID": id,
		},
	); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", id)
	}

	return nil
}

func (s *store) Unlock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": nil,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	return nil
}
