package mongodb

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userStore struct {
	usersCollection *mongo.Collection
}

func NewUserStore(database *mongo.Database) storage.UserStore {
	return &userStore{
		usersCollection: database.Collection("users"),
	}
}

func (u *userStore) CreateUser(username string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate username
	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username": username,
		},
	)
	if result.Err() == nil {
		return "", errors.Errorf(
			"a user with the username %q already exists",
			username,
		)
	} else if result.Err() != mongo.ErrNoDocuments {
		return "", errors.Wrapf(
			result.Err(),
			"error checking for existing user with the username %q",
			username,
		)
	}

	id := uuid.NewV4().String()

	if _, err :=
		u.usersCollection.InsertOne(
			ctx,
			bson.M{
				"id":       id,
				"username": username,
			},
		); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating user %q",
			username,
		)
	}
	return id, nil
}

func (u *userStore) GetUsers() ([]*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := u.usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users")
	}
	users := []*brignext.User{}
	for cur.Next(ctx) {
		user := &brignext.User{}
		err := cur.Decode(user)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding users")
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *userStore) GetUser(id string) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving user %q",
			id,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user %q",
			id,
		)
	}
	return user, nil
}

func (u *userStore) GetUserByUsername(username string) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username": username,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving user with username %q",
			username,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user with username %q",
			username,
		)
	}
	return user, nil
}

func (u *userStore) DeleteUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.DeleteOne(
			ctx,
			bson.M{
				"username": username,
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error deleting user %q",
			username,
		)
	}
	return nil
}
