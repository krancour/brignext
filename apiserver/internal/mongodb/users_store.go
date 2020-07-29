package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/users"
	errs "github.com/krancour/brignext/v2/internal/errors"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type usersStore struct {
	*BaseStore
	collection         *mongo.Collection
	sessionsCollection *mongo.Collection
}

func NewUsersStore(database *mongo.Database) (users.Store, error) {
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
	return &usersStore{
		BaseStore: &BaseStore{
			Database: database,
		},
		collection:         collection,
		sessionsCollection: database.Collection("sessions"),
	}, nil
}

func (u *usersStore) Create(ctx context.Context, user brignext.User) error {
	now := time.Now()
	user.Created = &now
	if _, err :=
		u.collection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return errs.NewErrConflict(
					"User",
					user.ID,
					fmt.Sprintf("A user with the ID %q already exists.", user.ID),
				)
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (u *usersStore) List(ctx context.Context) (brignext.UserList, error) {
	userList := brignext.NewUserList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	cur, err := u.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return userList, errors.Wrap(err, "error finding users")
	}
	if err := cur.All(ctx, &userList.Items); err != nil {
		return userList, errors.Wrap(err, "error decoding users")
	}
	return userList, nil
}

func (u *usersStore) Get(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user := brignext.User{}
	res := u.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, errs.NewErrNotFound("User", id)
	}
	if res.Err() != nil {
		return user, errors.Wrapf(res.Err(), "error finding user %q", id)
	}
	if err := res.Decode(&user); err != nil {
		return user, errors.Wrapf(err, "error decoding user %q", id)
	}
	return user, nil
}

func (u *usersStore) Lock(ctx context.Context, id string) error {
	res, err := u.collection.UpdateOne(
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
		return errs.NewErrNotFound("User", id)
	}

	// Now delete all the user's sessions. Note we're deliberately not doing this
	// in a transaction. This way if an error occurs after successfully locking
	// the user, but BEFORE OR WHILE deleting their existing sessions, at least
	// the user will be locked.
	if _, err = u.sessionsCollection.DeleteMany(
		ctx,
		bson.M{
			"userID": id,
		},
	); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", id)
	}

	return nil
}

func (u *usersStore) Unlock(ctx context.Context, id string) error {
	res, err := u.collection.UpdateOne(
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
		return errs.NewErrNotFound("User", id)
	}
	return nil
}